package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/moia-oss/aws-cfg-generator/pkg/cmd"
	"github.com/moia-oss/aws-cfg-generator/pkg/gen"
)

func setup(configFileContents string) (filename string) {
	file, err := ioutil.TempFile("", "aws-config")
	if err != nil {
		log.Fatal(err)
	}

	_, err = file.WriteString(configFileContents)
	if err != nil {
		log.Fatal(err)
	}

	return file.Name()
}

type TestCase struct {
	describe       string
	it             string
	originalConfig string
	expectedConfig string
	run            func(filename string)
}

func TestAll(t *testing.T) {
	roleArns := []string{"arn:aws:iam::12345:role/my-role"}
	accountMap := map[string]string{"12345": "my-account"}

	testCases := []TestCase{
		{
			describe:       "vault",
			it:             "generates a basic profile",
			originalConfig: `[default]`,
			expectedConfig: `[default]

[profile my-account]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				})
			},
		},
		{
			describe:       "vault",
			it:             "skips invalid roles",
			originalConfig: `[default]`,
			expectedConfig: `[default]

[profile my-account]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, []string{"foobar", "arn:aws:iam::12345:role/my-role"}, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				})
			},
		},
		{
			describe:       "vault",
			it:             "falls back to account numbers",
			originalConfig: `[default]`,
			expectedConfig: `[default]

[profile 67890]
role_arn        = arn:aws:iam::67890:role/my-role
source_profile  = default
include_profile = default
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, []string{"arn:aws:iam::67890:role/my-role"}, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				})
			},
		},
		{
			describe:       "vault",
			it:             "accepts a custom source profile",
			originalConfig: "[profile my-profile]",
			expectedConfig: `[profile my-profile]

[profile my-account]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = my-profile
include_profile = my-profile
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        "my-profile",
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				})
			},
		},
		{
			describe:       "vault",
			it:             "allows generation of profiles with role names",
			originalConfig: `[default]`,
			expectedConfig: `[default]

[profile my-account_my-role]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "",
				})
			},
		},
		{
			describe: "vault",
			it:       "keeps custom config options if set to true",
			originalConfig: `[default]
output = json

[profile my-account_my-role]
output          = json
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default

[profile some-other-profile]
`,
			expectedConfig: `[default]
output = json

[profile my-account_my-role]
output          = json
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default

[profile some-other-profile]
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     true,
					UseRoleNameInProfile: true,
					Region:               "",
				})
			},
		},
		{
			describe: "vault",
			it:       "deletes custom config options (but retains the source profile) if set to false",
			originalConfig: `[default]
output = json

[profile my-account_my-role]
output          = json
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default

[profile some-other-profile]
`,
			expectedConfig: `[default]
output = json

[profile my-account_my-role]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "",
				})
			},
		},
		{
			describe:       "vault",
			it:             "sets the region if supplied",
			originalConfig: `[default]`,
			expectedConfig: `[default]

[profile my-account_my-role]
role_arn        = arn:aws:iam::12345:role/my-role
source_profile  = default
include_profile = default
region          = eu-central-1
`,
			run: func(filename string) {
				gen.GenerateVaultProfile(accountMap, roleArns, cmd.VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "eu-central-1",
				})
			},
		},
		{
			describe:       "switch-roles",
			it:             "generates a basic profile with colors",
			originalConfig: ``,
			expectedConfig: `[my-account]
aws_account_id = 12345
role_name      = my-role
color          = ffffff
`,
			run: func(filename string) {
				gen.GenerateSwitchRolesProfile(accountMap, roleArns, cmd.SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: false,
					Color:                "ffffff",
				})
			},
		},
		{
			describe:       "switch-roles",
			it:             "uses role names",
			originalConfig: ``,
			expectedConfig: `[my-account_my-role]
aws_account_id = 12345
role_name      = my-role
color          = ffffff
`,
			run: func(filename string) {
				gen.GenerateSwitchRolesProfile(accountMap, roleArns, cmd.SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: true,
					Color:                "ffffff",
				})
			},
		},
		{
			describe:       "switch-roles",
			it:             "falls back to account numbers",
			originalConfig: ``,
			expectedConfig: `[67890]
aws_account_id = 67890
role_name      = my-role
color          = ffffff
`,
			run: func(filename string) {
				gen.GenerateSwitchRolesProfile(accountMap, []string{"arn:aws:iam::67890:role/my-role"}, cmd.SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: false,
					Color:                "ffffff",
				})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s %s", testCase.describe, testCase.it), func(t *testing.T) {

			filename := setup(testCase.originalConfig)
			defer os.Remove(filename)

			testCase.run(filename)

			actualConfig := getFile(filename)

			if testCase.expectedConfig != actualConfig {
				t.Errorf(`Expected
--------
%s
Got
--------
%s`, testCase.expectedConfig, actualConfig)
			}
		})
	}
}

func getFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return strings.TrimSuffix(string(content), "\n")
}
