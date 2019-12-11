# mackerel-plugin-aws-transitgateway

## use Assume Role

create IAM Role with the AWS Account that created Transit Gateway.

- no MFA
- allowed Policy
    - CloudWatchReadOnlyAccess

create IAM Policy with the AWS Account that runs mackerel-plugin-aws-transitgateway

```json
{
    "Version": "2012-10-17",
    "Statement": {
        "Effect": "Allow",
        "Action": "sts:AssumeRole",
        "Resource": "arn:aws:iam::123456789012:role/YourIAMRoleName"
    }
}
```

attach IAM Policy to AWS Resouce that runs mackerel-plugin-aws-transitgateway

## Synopsis

use assume role
```shell
mackerel-plugin-aws-transitgateway -role-arn <IAM Role Arn> -region <region> -tgw <Transit Gateway Resource ID>
```

use aws profile
```shell
mackerel-plugin-aws-transitgateway -region <region> -tgw <Transit Gateway Resource ID> [-access-key-id <AWS Access Key ID> -secret-key-id <WS Secret Access Key ID>]
```
