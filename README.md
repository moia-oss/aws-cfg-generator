# aws-cfg-generator

aws-cfg-generator is a CLI tool to generate configs for AWS helper tools based on an IAM user's permissions.

To use this tool you need AWS credentials for an IAM user.

## Supported tools

aws-cfg-generator can generate a config for [aws-vault](https://github.com/99designs/aws-vault) and 
[aws-extend-switch-roles](https://github.com/tilfinltd/aws-extend-switch-roles)

Note that technically the aws-cfg-generator does not depend on aws-vault, but to run it does require AWS credentials
of an IAM user to be present somewhere in the credentials provider chain. In the rest of this documentation, we assume 
that aws-vault is used to supply these credentials.

### aws-vault

First, add a profile named `default` to your machine with `aws-vault add default` then put in your aws-access-key and
aws-secret-access-key as prompted (you can generate an access key in the AWS web console under "My Security 
Credentials").

Then simply run `aws-vault exec default -- ./aws-cfg-generator vault >> ~/.aws/config` to add a profile to 
aws-vault for every profile you're explicitly allowed to assume. Run `cat ~/.aws/config` to verify it worked. The config
should look something like this:

```
[profile 123456789098]
region=eu-central-1
role_arn=arn:aws:iam::123456789098:role/role-name
source_profile=default
mfa_serial=arn:aws:iam::111111111111:mfa/user-name@example.com

[profile 098765432123]
region=eu-central-1
role_arn=arn:aws:iam::098765432123:role/role-name-two
source_profile=default
mfa_serial=arn:aws:iam::111111111111:mfa/user-name@example.com

. . .
```

#### Flags

- `source-profile` Can be used to specify a source profile other than `default`
- `region` Can be used to specify a region other than `eu-central-1`


### aws-extend-switch-roles

Run `aws-vault exec default -- ./aws-cfg-generator switch-roles` to write your config to standard out. Then copy/paste
it into your aws-extend-switch-roles settings page. The generated config should look something like this:

```
[123456789098]
aws_account_id = 123456789098
role_name = example-role
color = 00ff7f

[098765432123]
aws_account_id = 098765432123
role_name = example-role-two
color = 00ff7f

[. . .]
```

#### Flags

- `color` Specify what color you want to represent this profile in the web console as a hex value

## Known-limitations

- No human-readable profile names
- Only recognizes policies that are attached to groups
- Can only recognize explicit permissions (i.e. it doesn't work when the `Resource` is not a role ARN)
- Assumes there's an MFA
  
## Planned features

- Human-readable names configurable via a provided mapping from AWS account IDs to names
- Discover roles that are attached or inlined directly on the user

## Contributions

Contributions are welcome! Just provide a well-documented PR and a maintainer will review it soon.

## Releases and building

A github release is automatically published for every new `v*` tag and provides binaries for darwin, windows, and 
linux. If your operating system or architecture isn't provided feel free to make a PR that adds it, or build it 
yourself by running `go build aws-cfg-generator/main.go`.
