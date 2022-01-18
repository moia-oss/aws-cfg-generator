# aws-cfg-generator

aws-cfg-generator is a CLI tool to generate configs for AWS helper tools based on an IAM user's permissions.

To use this tool you need AWS credentials for an IAM user. This IAM user also needs sufficient permissions to read their
own permission sets and group memberships.

```
Usage: aws-cfg-generator <command>

Flags:
  -h, --help    Show context-sensitive help.

Commands:
  vault --vault-config-path=STRING
    generates a config for aws-vault

  switch-roles --output-file=STRING
    generates a config for aws-extend-switch-roles
```

## Profile names

In order to name profiles correctly, aws-cfg-generator will attempt to call `organizations.ListAccounts` and match that
with account IDs in the roles the user has access to. If the user has permissions for a role not in the same AWS
organization the profile will be named by the account ID instead. Similarly, if the user lacks permissions to list the
organization's accounts, the profiles will be named by account IDs as well,

## Supported tools

aws-cfg-generator can generate a config for:

- [aws-vault](https://github.com/99designs/aws-vault)
- [aws-extend-switch-roles](https://github.com/tilfinltd/aws-extend-switch-roles)

Note that technically the aws-cfg-generator does not depend on aws-vault, but to run it does require AWS credentials
of an IAM user to be present somewhere in the credentials provider chain. In the rest of this documentation, we assume
that aws-vault is used to supply these credentials.

### aws-vault

```sh
CONFIG=$HOME/.aws/config

# First, add a profile named `default` to your machine with:
# Put in your aws-access-key and aws-secret-access-key as prompted
# (you can generate an access key in the AWS web console under "My Security Credentials").
aws-vault add default

# backup your config
cp ${CONFIG} ${CONFIG}.bak

# now run the command to add a profile to aws-vault for every profile you're explicitly allowed to assume
aws-vault exec default -- ./aws-cfg-generator vault --vault-config-path=${CONFIG}
# verify that it worked
cat ${CONFIG}
# delete the backup
rm ${CONFIG}.bak
```

The output should look like this:

```ini
[profile account-name]
role_arn=arn:aws:iam::123456789098:role/role-name
source_profile=default
include_profile=default

[profile another-account-name]
role_arn=arn:aws:iam::098765432123:role/role-name-two
source_profile=default
include_profile=default

# ...
```

#### Flags

```
REQUIRED

--vault-config-path=STRING         Where to load/save the config

OPTIONAL

--source-profile="default"         The profile that your credentials should come from
--region=STRING                    Override the region configured with your source profile
--keep-custom-config=true          Retains any custom profiles or settings. Set to false to remove everything
                                   except the source profile and generated config
--use-role-name-in-profile=false   Append the role name to the profile name
--role=STRING                      If set, then a profile with this role will be generated for every account in the organization, in addition to the roles that the user has permissions to assume
```

Note: When using the `--role` flag we do not check to see if the user has permission to assume that role. This is useful
if the user has a policy that allows them e.g. `sts:AssumeRole` on resource `*` and the target accounts
manage who is allowed to assume various roles.

### aws-extend-switch-roles

Run `aws-vault exec default -- ./aws-cfg-generator switch-roles --output-file=output.ini`, then copy/paste it into your aws-extend-switch-roles settings page.

The generated config should look something like this:

```ini
[account-name]
aws_account_id = 123456789098
role_name = example-role
color = 00ff7f

[another-account-name]
aws_account_id = 098765432123
role_name = example-role-two
color = 00ff7f

# ...
```

#### Flags

```
REQUIRED

--output-file=STRING                Where to save the config.

OPTIONAL

--color="00ff7f"                    The hexcode color that should be set for each profile
--prd-color="ff0000"                The hexcode color that should be set for each profile which name ends in 'prd' or 'global'
--use-role-name-in-profile=false    Append the role name to the profile name
```

## Known-limitations

- Only recognizes policies that are attached to groups
- Can only recognize explicit permissions (i.e. it doesn't work when the `Resource` is not a role ARN)

## Planned features

- Discover roles that are attached or inlined directly on the user

## Contributions

Contributions are welcome! Just provide a well-documented PR and a maintainer will review it as soon as possible.

## Releases and building

A GitHub release is automatically published for every new `v*` tag and provides binaries for darwin, windows, and
linux. If your operating system or architecture isn't provided feel free to make a PR that adds it, or build it
yourself by running `go build aws-cfg-generator/main.go`.
