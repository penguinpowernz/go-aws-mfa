# go-aws-mfa

A tool to help with using AWS CLI when MFA is enabled.

## Usage

Assuming the AWS best practices are used, with a privilege-less master account that is logged into, and then a role is assumed for accessing each account specific environment, you would setup your config like this:

```
[default-long-term]
aws_secret_access_key = YOUR_MASTER_KEY
aws_access_key_id     = YOUR_MASTER_ID
aws_mfa_device        = YOUR_MASTER_MFA_DEVICE_ARN

[prod]
long_term             = default
assume_role           = arn:aws:iam::1234567890:role/AssumedRole

[dev]
long_term             = default
assume_role           = arn:aws:iam::9999999999:role/AssumedRole
```

By setting up your `~/.aws/credentials` file with this pattern you can quickly and easily assume roles and authenticate using an MFA code.

    $ go-aws-mfa prod
    Authenticating for prod
    Sourcing creds from default-long-term
    Assuming role arn:aws:iam::1234567890:role/AssumedRole
    Using the MFA device arn:aws:iam::0987654321:mfa/username
    Enter MFA code: 748871
    Credentials updated for prod, valid until 2019-11-18 19:45:57 +0000 UTC

After this, the `[prod]` section of your config file would look like this:

```
[prod]
long_term             = default
assume_role           = arn:aws:iam::1234567890:role/AssumedRole
aws_access_key_id     = SHORT_TERM_ID
aws_secret_access_key = SHORT_TERM_KEY
aws_session_token     = SHORT_TERM_TOKEN
```

Note that this requires the `~/.aws/config` to not contain `role_arn` or `source_profile` or weird things can happen.  You can still keep region there though:

```
[profile prod]
region = intl-antarctica-01
```

Now you can use your short term creds like so:

    aws --profile prod cognito-idp list-user-pools

## Download

You can obtain the binary from the [releases](https://github.com/penguinpowernz/go-aws-mfa/releases) section.
