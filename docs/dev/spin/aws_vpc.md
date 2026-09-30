# AWS VPC

Generate a private VPC template and a CloudFormation parameters file. Supply the VPC CIDR as the argument and repeat `--subnet-cidr` for every subnet:

```sh
bam dev spin aws-vpc 10.40.0.0/16 \
  --subnet-cidr 10.40.1.0/24 \
  --subnet-cidr 10.40.2.0/24 \
  --ecr-api-endpoint \
  --ecr-docker-endpoint
```

By default, this writes `vpc.yaml` and `vpc-parameters.json`. Override the paths with `--template-file` and `--parameters-file`.

The template creates private subnets with public IP assignment disabled, isolated route tables, rejected-traffic VPC Flow Logs encrypted with a rotating KMS key, and a closed default security group. Endpoints are excluded by default. Opt in independently with `--ec2-api-endpoint`, `--ecr-api-endpoint`, `--ecr-docker-endpoint`, `--ssm-endpoint`, or `--s3-gateway-endpoint`; these selections are recorded in the parameters JSON and can also be changed at deployment time. Interface endpoint network interfaces are placed in the first generated subnet. Deploying the template requires `CAPABILITY_IAM` because it creates roles and a Lambda-backed custom resource.

```sh
aws cloudformation create-stack \
  --stack-name secure-vpc \
  --template-body file://vpc.yaml \
  --parameters file://vpc-parameters.json \
  --capabilities CAPABILITY_IAM
```

Interface endpoints can incur hourly and data processing charges. The S3 gateway endpoint has no hourly charge. Flow logging and the KMS key also have AWS charges. This template addresses VPC-level controls in the AWS Foundational Security Best Practices standard, but cannot guarantee a clean account-wide Security Hub view: account-level controls, organization policies, required tag keys, regional service availability, and other enabled standards can introduce additional findings.