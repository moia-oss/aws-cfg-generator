package cmd

/*
   Copyright 2021 MOIA GmbH
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func setup(configFileContents string) (filename string) {
	file, err := os.CreateTemp("", "aws-config")
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, []string{"foobar", "arn:aws:iam::12345:role/my-role"}, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, []string{"arn:aws:iam::67890:role/my-role"}, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        "my-profile",
					KeepCustomConfig:     false,
					UseRoleNameInProfile: false,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     true,
					UseRoleNameInProfile: true,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "",
				}, true)
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
				generateVaultProfile(accountMap, roleArns, VaultCmd{
					VaultConfigPath:      filename,
					SourceProfile:        `default`,
					KeepCustomConfig:     false,
					UseRoleNameInProfile: true,
					Region:               "eu-central-1",
				}, true)
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
				generateSwitchRolesProfile(accountMap, roleArns, SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: false,
					Color:                "ffffff",
				}, true)
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
				generateSwitchRolesProfile(accountMap, roleArns, SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: true,
					Color:                "ffffff",
				}, true)
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
				generateSwitchRolesProfile(accountMap, []string{"arn:aws:iam::67890:role/my-role"}, SwitchRolesCmd{
					OutputFile:           filename,
					UseRoleNameInProfile: false,
					Color:                "ffffff",
				}, true)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s %s", testCase.describe, testCase.it), func(t *testing.T) {
			filename := setup(testCase.originalConfig)
			defer func() {
				err := os.Remove(filename)
				if err != nil {
					log.Panic(err)
				}
			}()

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
	content, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return string(content)
}
