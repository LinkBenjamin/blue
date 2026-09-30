package spin

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	vpcCIDR              string
	subnetCIDRs          []string
	vpcParametersFile    string
	vpcTemplateFile      string
	addEc2ApiEndpoint    bool
	addEcrApiEndpoint    bool
	addEcrDockerEndpoint bool
	addSsmEndpoint       bool
	addS3GatewayEndpoint bool
)

var vpcCmd = &cobra.Command{
	Use:   "aws-vpc [vpc-cidr]",
	Short: "Generate a private VPC template with Security Hub baseline controls",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcPrefix, err := parseIPv4CIDR(args[0], "VPC CIDR")
		if err != nil {
			return err
		}
		if vpcPrefix.Bits() < 16 || vpcPrefix.Bits() > 28 {
			return fmt.Errorf("VPC CIDR prefix must be between /16 and /28")
		}
		if len(subnetCIDRs) == 0 {
			return fmt.Errorf("provide at least one subnet with --subnet-cidr")
		}
		if len(subnetCIDRs) > 100 {
			return fmt.Errorf("at most 100 subnets are supported")
		}

		subnetPrefixes := make([]netip.Prefix, len(subnetCIDRs))
		for index, cidr := range subnetCIDRs {
			subnetPrefixes[index], err = parseIPv4CIDR(cidr, fmt.Sprintf("subnet %d CIDR", index+1))
			if err != nil {
				return err
			}
			if subnetPrefixes[index].Bits() < 16 || subnetPrefixes[index].Bits() > 28 {
				return fmt.Errorf("subnet %d CIDR prefix must be between /16 and /28", index+1)
			}
			if !vpcPrefix.Contains(subnetPrefixes[index].Addr()) || subnetPrefixes[index].Bits() < vpcPrefix.Bits() {
				return fmt.Errorf("subnet CIDR %q is not contained by VPC CIDR %q", subnetPrefixes[index], vpcPrefix)
			}
			for previous := 0; previous < index; previous++ {
				if prefixesOverlap(subnetPrefixes[index], subnetPrefixes[previous]) {
					return fmt.Errorf("subnet CIDRs %q and %q overlap", subnetPrefixes[previous], subnetPrefixes[index])
				}
			}
		}
		if err := validateOutputPaths(vpcTemplateFile, vpcParametersFile); err != nil {
			return err
		}

		parameters := buildVPCParameters(vpcPrefix.String(), subnetPrefixes, vpcEndpointOptions{
			addEc2ApiEndpoint:    addEc2ApiEndpoint,
			addEcrApiEndpoint:    addEcrApiEndpoint,
			addEcrDockerEndpoint: addEcrDockerEndpoint,
			addSsmEndpoint:       addSsmEndpoint,
			addS3GatewayEndpoint: addS3GatewayEndpoint,
		})
		parameterJSON, err := json.MarshalIndent(parameters, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode VPC parameters: %w", err)
		}
		parameterJSON = append(parameterJSON, '\n')
		template := buildVPCTemplate(len(subnetPrefixes))
		if err := os.WriteFile(vpcTemplateFile, []byte(template), 0644); err != nil {
			return fmt.Errorf("failed to write VPC template %q: %w", vpcTemplateFile, err)
		}
		if err := os.WriteFile(vpcParametersFile, parameterJSON, 0644); err != nil {
			return fmt.Errorf("failed to write VPC parameters %q: %w", vpcParametersFile, err)
		}

		fmt.Printf("Wrote VPC template to %q and parameters to %q\n", vpcTemplateFile, vpcParametersFile)
		return nil
	},
}

func init() {
	vpcCmd.Flags().StringArrayVar(&subnetCIDRs, "subnet-cidr", nil, "IPv4 CIDR for a subnet (repeat for each subnet)")
	vpcCmd.Flags().StringVar(&vpcParametersFile, "parameters-file", "vpc-parameters.json", "Path to write the CloudFormation parameters JSON file")
	vpcCmd.Flags().StringVar(&vpcTemplateFile, "template-file", "vpc.yaml", "Path to write the CloudFormation template")
	vpcCmd.Flags().BoolVar(&addEc2ApiEndpoint, "ec2-api-endpoint", false, "Create the EC2 API interface endpoint")
	vpcCmd.Flags().BoolVar(&addEcrApiEndpoint, "ecr-api-endpoint", false, "Create the ECR API interface endpoint")
	vpcCmd.Flags().BoolVar(&addEcrDockerEndpoint, "ecr-docker-endpoint", false, "Create the ECR Docker Registry interface endpoint")
	vpcCmd.Flags().BoolVar(&addSsmEndpoint, "ssm-endpoint", false, "Create the Systems Manager interface endpoint")
	vpcCmd.Flags().BoolVar(&addS3GatewayEndpoint, "s3-gateway-endpoint", false, "Create the S3 gateway endpoint")
}

type cloudFormationParameter struct {
	ParameterKey   string `json:"ParameterKey"`
	ParameterValue string `json:"ParameterValue"`
}

type vpcEndpointOptions struct {
	addEc2ApiEndpoint    bool
	addEcrApiEndpoint    bool
	addEcrDockerEndpoint bool
	addSsmEndpoint       bool
	addS3GatewayEndpoint bool
}

func buildVPCParameters(vpcCIDR string, subnetCIDRs []netip.Prefix, endpoints vpcEndpointOptions) []cloudFormationParameter {
	parameters := []cloudFormationParameter{{ParameterKey: "VpcCidr", ParameterValue: vpcCIDR}}
	for index, prefix := range subnetCIDRs {
		parameters = append(parameters, cloudFormationParameter{
			ParameterKey:   fmt.Sprintf("Subnet%02dCidr", index+1),
			ParameterValue: prefix.String(),
		})
	}
	parameters = append(parameters,
		cloudFormationParameter{ParameterKey: "AddEc2ApiEndpoint", ParameterValue: strconv.FormatBool(endpoints.addEc2ApiEndpoint)},
		cloudFormationParameter{ParameterKey: "AddEcrApiEndpoint", ParameterValue: strconv.FormatBool(endpoints.addEcrApiEndpoint)},
		cloudFormationParameter{ParameterKey: "AddEcrDockerEndpoint", ParameterValue: strconv.FormatBool(endpoints.addEcrDockerEndpoint)},
		cloudFormationParameter{ParameterKey: "AddSsmEndpoint", ParameterValue: strconv.FormatBool(endpoints.addSsmEndpoint)},
		cloudFormationParameter{ParameterKey: "AddS3GatewayEndpoint", ParameterValue: strconv.FormatBool(endpoints.addS3GatewayEndpoint)},
	)
	return parameters
}

func parseIPv4CIDR(value, label string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil || !prefix.Addr().Is4() {
		return netip.Prefix{}, fmt.Errorf("%s %q must be a valid IPv4 CIDR", label, value)
	}
	return prefix.Masked(), nil
}

func prefixesOverlap(left, right netip.Prefix) bool {
	return left.Contains(right.Addr()) || right.Contains(left.Addr())
}

func validateOutputPaths(templatePath, parametersPath string) error {
	if templatePath == parametersPath {
		return fmt.Errorf("template and parameters output paths must be different")
	}
	return nil
}

func buildVPCTemplate(subnetCount int) string {
	var template strings.Builder
	template.WriteString(`AWSTemplateFormatVersion: '2010-09-09'
Description: Private VPC with flow logging, isolated subnets, closed default security group, and private service endpoints.

Parameters:
  VpcCidr:
    Type: String
    Description: IPv4 CIDR block for the VPC
`)
	for index := 1; index <= subnetCount; index++ {
		fmt.Fprintf(&template, `  Subnet%02dCidr:
    Type: String
    Description: IPv4 CIDR block for subnet %02d
`, index, index)
	}
	template.WriteString(`  AddEc2ApiEndpoint:
    Type: String
    AllowedValues: ['true', 'false']
    Default: 'false'
    Description: Create the EC2 API interface endpoint
  AddEcrApiEndpoint:
    Type: String
    AllowedValues: ['true', 'false']
    Default: 'false'
    Description: Create the ECR API interface endpoint
  AddEcrDockerEndpoint:
    Type: String
    AllowedValues: ['true', 'false']
    Default: 'false'
    Description: Create the ECR Docker Registry interface endpoint
  AddSsmEndpoint:
    Type: String
    AllowedValues: ['true', 'false']
    Default: 'false'
    Description: Create the Systems Manager interface endpoint
  AddS3GatewayEndpoint:
    Type: String
    AllowedValues: ['true', 'false']
    Default: 'false'
    Description: Create the S3 gateway endpoint

Conditions:
  CreateEc2ApiEndpoint: !Equals [!Ref AddEc2ApiEndpoint, 'true']
  CreateEcrApiEndpoint: !Equals [!Ref AddEcrApiEndpoint, 'true']
  CreateEcrDockerEndpoint: !Equals [!Ref AddEcrDockerEndpoint, 'true']
  CreateSsmEndpoint: !Equals [!Ref AddSsmEndpoint, 'true']
  CreateS3GatewayEndpoint: !Equals [!Ref AddS3GatewayEndpoint, 'true']
  HasInterfaceEndpoint: !Or
    - !Condition CreateEc2ApiEndpoint
    - !Condition CreateEcrApiEndpoint
    - !Condition CreateEcrDockerEndpoint
    - !Condition CreateSsmEndpoint

`)
	template.WriteString(`
Resources:
  Vpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref VpcCidr
      EnableDnsSupport: true
      EnableDnsHostnames: true
      Tags:
        - Key: Name
          Value: secure-vpc

`)
	for index := 1; index <= subnetCount; index++ {
		fmt.Fprintf(&template, `  Subnet%02d:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref Vpc
      CidrBlock: !Ref Subnet%02dCidr
      MapPublicIpOnLaunch: false
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-private-%02d'

`, index, index, index)
		fmt.Fprintf(&template, `  Subnet%02dRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref Vpc
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-private-rt-%02d'

  Subnet%02dRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref Subnet%02d
      RouteTableId: !Ref Subnet%02dRouteTable

`, index, index, index, index, index)
	}
	template.WriteString(`  FlowLogsKey:
    Type: AWS::KMS::Key
    Properties:
      EnableKeyRotation: true
      KeyPolicy:
        Version: '2012-10-17'
        Statement:
          - Sid: EnableAccountIAMPermissions
            Effect: Allow
            Principal:
              AWS: !Sub 'arn:${AWS::Partition}:iam::${AWS::AccountId}:root'
            Action: kms:*
            Resource: '*'
          - Sid: AllowCloudWatchLogsEncryption
            Effect: Allow
            Principal:
              Service: !Sub 'logs.${AWS::Region}.${AWS::URLSuffix}'
            Action:
              - kms:Encrypt*
              - kms:Decrypt*
              - kms:ReEncrypt*
              - kms:GenerateDataKey*
              - kms:Describe*
            Resource: '*'
            Condition:
              ArnLike:
                kms:EncryptionContext:aws:logs:arn: !Sub 'arn:${AWS::Partition}:logs:${AWS::Region}:${AWS::AccountId}:log-group:/aws/vpc/${AWS::StackName}/flow-logs'

  FlowLogGroup:
    Type: AWS::Logs::LogGroup
    Properties:
      LogGroupName: !Sub '/aws/vpc/${AWS::StackName}/flow-logs'
      KmsKeyId: !GetAtt FlowLogsKey.Arn
      RetentionInDays: 365
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-vpc-flow-logs'

  FlowLogRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: vpc-flow-logs.amazonaws.com
            Action: sts:AssumeRole
      Policies:
        - PolicyName: PublishVpcFlowLogs
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action: logs:DescribeLogGroups
                Resource: '*'
              - Effect: Allow
                Action:
                  - logs:CreateLogStream
                  - logs:PutLogEvents
                  - logs:DescribeLogStreams
                Resource: !Sub '${FlowLogGroup.Arn}:*'

  VpcFlowLog:
    Type: AWS::EC2::FlowLog
    Properties:
      ResourceId: !Ref Vpc
      ResourceType: VPC
      TrafficType: REJECT
      LogDestinationType: cloud-watch-logs
      LogGroupName: !Ref FlowLogGroup
      DeliverLogsPermissionArn: !GetAtt FlowLogRole.Arn
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-vpc-rejected-traffic'

  DefaultSecurityGroupCleanerRole:
    Type: AWS::IAM::Role
    Properties:
      AssumeRolePolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Effect: Allow
            Principal:
              Service: lambda.amazonaws.com
            Action: sts:AssumeRole
      Policies:
        - PolicyName: CloseDefaultSecurityGroup
          PolicyDocument:
            Version: '2012-10-17'
            Statement:
              - Effect: Allow
                Action: ec2:DescribeSecurityGroups
                Resource: '*'
              - Effect: Allow
                Action:
                  - ec2:RevokeSecurityGroupIngress
                  - ec2:RevokeSecurityGroupEgress
                Resource: !Sub 'arn:${AWS::Partition}:ec2:${AWS::Region}:${AWS::AccountId}:security-group/*'

  DefaultSecurityGroupCleaner:
    Type: AWS::Lambda::Function
    Properties:
      Runtime: python3.12
      Handler: index.handler
      Timeout: 60
      Role: !GetAtt DefaultSecurityGroupCleanerRole.Arn
      Code:
        ZipFile: |
          import boto3
          import json
          import urllib.request

          def send(event, context, status, physical_id, reason=''):
              body = json.dumps({
                  'Status': status,
                  'Reason': reason[:1000] or 'See CloudWatch Logs for details',
                  'PhysicalResourceId': physical_id,
                  'StackId': event['StackId'],
                  'RequestId': event['RequestId'],
                  'LogicalResourceId': event['LogicalResourceId'],
                  'NoEcho': False,
                  'Data': {}
              }).encode()
              request = urllib.request.Request(event['ResponseURL'], data=body, method='PUT', headers={'content-type': ''})
              urllib.request.urlopen(request)

          def handler(event, context):
              physical_id = event.get('PhysicalResourceId', 'default-security-group-cleaner')
              try:
                  if event['RequestType'] != 'Delete':
                      group_id = event['ResourceProperties']['GroupId']
                      ec2 = boto3.client('ec2')
                      group = ec2.describe_security_groups(GroupIds=[group_id])['SecurityGroups'][0]
                      for field, revoke in (('IpPermissions', ec2.revoke_security_group_ingress), ('IpPermissionsEgress', ec2.revoke_security_group_egress)):
                          permissions = group.get(field, [])
                          if permissions:
                              revoke(GroupId=group_id, IpPermissions=permissions)
                      physical_id = group_id
                  send(event, context, 'SUCCESS', physical_id)
              except Exception as error:
                  send(event, context, 'FAILED', physical_id, str(error))

  CloseDefaultSecurityGroup:
    Type: Custom::CloseDefaultSecurityGroup
    Properties:
      ServiceToken: !GetAtt DefaultSecurityGroupCleaner.Arn
      GroupId: !GetAtt Vpc.DefaultSecurityGroup

  EndpointSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Condition: HasInterfaceEndpoint
    Properties:
      GroupDescription: HTTPS access to private AWS service endpoints from this VPC
      VpcId: !Ref Vpc
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: !Ref VpcCidr
      SecurityGroupEgress: []
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-endpoint-https'

  Ec2ApiEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Condition: CreateEc2ApiEndpoint
    Properties:
      VpcId: !Ref Vpc
      VpcEndpointType: Interface
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.ec2'
      PrivateDnsEnabled: true
      SubnetIds:
  `)
	writeEndpointSubnetReference(&template)
	template.WriteString(`      SecurityGroupIds:
        - !Ref EndpointSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-ec2-api'

  EcrApiEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Condition: CreateEcrApiEndpoint
    Properties:
      VpcId: !Ref Vpc
      VpcEndpointType: Interface
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.ecr.api'
      PrivateDnsEnabled: true
      SubnetIds:
  `)
	writeEndpointSubnetReference(&template)
	template.WriteString(`      SecurityGroupIds:
        - !Ref EndpointSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-ecr-api'

  EcrDockerEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Condition: CreateEcrDockerEndpoint
    Properties:
      VpcId: !Ref Vpc
      VpcEndpointType: Interface
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.ecr.dkr'
      PrivateDnsEnabled: true
      SubnetIds:
  `)
	writeEndpointSubnetReference(&template)
	template.WriteString(`      SecurityGroupIds:
        - !Ref EndpointSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-ecr-dkr'

  SystemsManagerEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Condition: CreateSsmEndpoint
    Properties:
      VpcId: !Ref Vpc
      VpcEndpointType: Interface
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.ssm'
      PrivateDnsEnabled: true
      SubnetIds:
  `)
	writeEndpointSubnetReference(&template)
	template.WriteString(`      SecurityGroupIds:
        - !Ref EndpointSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-ssm'

  S3GatewayEndpoint:
    Type: AWS::EC2::VPCEndpoint
    Condition: CreateS3GatewayEndpoint
    Properties:
      VpcId: !Ref Vpc
      VpcEndpointType: Gateway
      ServiceName: !Sub 'com.amazonaws.${AWS::Region}.s3'
      RouteTableIds:
`)
	for index := 1; index <= subnetCount; index++ {
		fmt.Fprintf(&template, "        - !Ref Subnet%02dRouteTable\n", index)
	}
	template.WriteString(`      Tags:
        - Key: Name
          Value: !Sub '${AWS::StackName}-s3-gateway'

Outputs:
  VpcId:
    Description: ID of the private VPC
    Value: !Ref Vpc
  VpcCidr:
    Description: IPv4 CIDR block of the VPC
    Value: !GetAtt Vpc.CidrBlock
`)
	for index := 1; index <= subnetCount; index++ {
		fmt.Fprintf(&template, "  Subnet%02dId:\n    Description: ID of private subnet %02d\n    Value: !Ref Subnet%02d\n", index, index, index)
	}
	return template.String()
}

func writeEndpointSubnetReference(template *strings.Builder) {
	template.WriteString("        - !Ref Subnet01\n")
}
