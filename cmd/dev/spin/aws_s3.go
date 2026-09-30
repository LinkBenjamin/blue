package spin

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Default YAML template format with dynamic options
const secureS3TemplateFormat = `AWSTemplateFormatVersion: '2010-09-09'
Description: S3 bucket compliant with AWS SecurityHub Foundational Security Best Practices.

Parameters:
  BucketName:
    Type: String
    Description: Name portion of the S3 bucket name
  BucketPrefix:
    Type: String
    Default: bucket-
    Description: Prefix to prepend to the bucket name
  NoncurrentDays:
    Type: Number
    Default: 90
    MinValue: 1
    Description: Days before expiring non-current object versions
  AbortMultipartDays:
    Type: Number
    Default: 7
    MinValue: 1
    Description: Days before aborting incomplete multipart uploads

Resources:
  SecureBucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: !Sub '${BucketPrefix}${BucketName}'
      # S3.1: Enforce Versioning
      VersioningConfiguration:
        Status: Enabled
      # S3.2 & S3.3: Block all public access
      PublicAccessBlockConfiguration:
        BlockPublicAcls: true
        BlockPublicPolicy: true
        IgnorePublicAcls: true
        RestrictPublicBuckets: true
      # S3.4: Server-side encryption enabled
      BucketEncryption:
        ServerSideEncryptionConfiguration:
          - ServerSideEncryptionByDefault:
              SSEAlgorithm: AES256
      # S3.13 / Config Rule s3-version-lifecycle-policy-check
      LifecycleConfiguration:
        Rules:
          - Id: SecurityHubLifecyclePolicy
            Status: Enabled
            NoncurrentVersionExpiration:
              NoncurrentDays: !Ref NoncurrentDays
            AbortIncompleteMultipartUpload:
              DaysAfterInitiation: !Ref AbortMultipartDays

  # S3.5: Enforce HTTPS-only traffic via Bucket Policy
  SecureBucketPolicy:
    Type: AWS::S3::BucketPolicy
    Properties:
      Bucket: !Ref SecureBucket
      PolicyDocument:
        Version: '2012-10-17'
        Statement:
          - Sid: AllowSSLRequestsOnly
            Effect: Deny
            Principal: '*'
            Action: 's3:*'
            Resource:
              - !GetAtt SecureBucket.Arn
              - !Sub '${SecureBucket.Arn}/*'
            Condition:
              Bool:
                'aws:SecureTransport': 'false'

Outputs:
  BucketName:
    Description: Name of the secure S3 Bucket
    Value: !Ref SecureBucket
  BucketArn:
    Description: ARN of the secure S3 Bucket
    Value: !GetAtt SecureBucket.Arn
`

var (
	prefix             string
	noncurrentDays     int
	abortMultipartDays int
	parametersFile     string
  templateFile       string
)

var s3Cmd = &cobra.Command{
	Use:   "aws-s3-bucket [bucket-name]",
	Short: "Generate a SecurityHub-compliant CloudFormation template for an S3 bucket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawBucketName := strings.TrimSpace(args[0])
		if rawBucketName == "" {
			return fmt.Errorf("bucket name cannot be empty")
		}

		parameters := []struct {
			ParameterKey   string `json:"ParameterKey"`
			ParameterValue string `json:"ParameterValue"`
		}{
			{ParameterKey: "BucketName", ParameterValue: rawBucketName},
			{ParameterKey: "BucketPrefix", ParameterValue: prefix},
			{ParameterKey: "NoncurrentDays", ParameterValue: strconv.Itoa(noncurrentDays)},
			{ParameterKey: "AbortMultipartDays", ParameterValue: strconv.Itoa(abortMultipartDays)},
		}
		parameterJSON, err := json.MarshalIndent(parameters, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode CloudFormation parameters: %w", err)
		}
		parameterJSON = append(parameterJSON, '\n')
		if err := os.WriteFile(parametersFile, parameterJSON, 0644); err != nil {
			return fmt.Errorf("failed to write parameters file %q: %w", parametersFile, err)
		}
    if err := os.WriteFile(templateFile, []byte(secureS3TemplateFormat), 0644); err != nil {
      return fmt.Errorf("failed to write CloudFormation template %q: %w", templateFile, err)
    }

    fmt.Printf("Wrote CloudFormation template to %q and parameters to %q\n", templateFile, parametersFile)
		return nil
	},
}

func init() {
	s3Cmd.Flags().StringVarP(&prefix, "prefix", "p", "bucket-", "Prefix to prepend to the bucket name")
	s3Cmd.Flags().IntVar(&noncurrentDays, "noncurrent-days", 90, "Number of days before expiring non-current object versions")
	s3Cmd.Flags().IntVar(&abortMultipartDays, "abort-multipart-days", 7, "Number of days before aborting incomplete multipart uploads")
  s3Cmd.Flags().StringVar(&parametersFile, "parameters-file", "s3-parameters.json", "Path to write the CloudFormation parameters JSON file")
  s3Cmd.Flags().StringVar(&templateFile, "template-file", "s3.yaml", "Path to write the CloudFormation template")
}
