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

	"github.com/moia-oss/aws-cfg-generator/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

// nolint:govet // we need the bare `required` tag here
type VaultCmd struct {
	SourceProfile        string `help:"The profile that your credentials should come from" default:"default"`
	Region               string `help:"Override the region configured with your source profile"`
	VaultConfigPath      string `help:"Where to load/save the config" required`
	KeepCustomConfig     bool   `help:"Retains any custom profiles or settings. Set to false to remove everything except the source profile and generated config" default:true`
	UseRoleNameInProfile bool   `help:"Append the role name to the profile name" default:false`
}

func (vc *VaultCmd) Run(cli *CLI) error {
	roleArns, accountMap := util.GetAWSContext().GetRolesAndAccounts()
	generateVaultProfile(accountMap, roleArns, cli.Vault)

	return nil
}

func generateVaultProfile(accountMap map[string]string, roleArns []string, cmdOptions VaultCmd) {
	config, err := ini.Load(cmdOptions.VaultConfigPath)
	if err != nil {
		log.Panic().Err(err).Str("file-path", cmdOptions.VaultConfigPath).Msg("could not load config")
	}

	sourceProfileSectionName := cmdOptions.SourceProfile

	// the profile can either be [default] or [profile foo]
	if sourceProfileSectionName != "default" {
		sourceProfileSectionName = fmt.Sprint("profile ", sourceProfileSectionName)
	}

	// make sure the source section exists
	_, err = config.GetSection(sourceProfileSectionName)
	if err != nil {
		log.Panic().Err(err).Str("section", sourceProfileSectionName).Msg("source profile not found")
	}

	// only copy the source profile and generated profiles, discard the rest of the config
	if !cmdOptions.KeepCustomConfig {
		newConfig := ini.Empty()

		setProfileKey := util.GetKeySetter(newConfig.Section(sourceProfileSectionName))

		for key, value := range config.Section(sourceProfileSectionName).KeysHash() {
			setProfileKey(key, value)
		}

		config = newConfig
	}

	for _, profile := range util.GetProfiles("profile ", accountMap, roleArns, cmdOptions.UseRoleNameInProfile) {
		profileSection := config.Section(profile.ProfileName)

		setKey := util.GetKeySetter(profileSection)

		setKey("role_arn", profile.RoleArn)
		setKey("source_profile", cmdOptions.SourceProfile)
		setKey("include_profile", cmdOptions.SourceProfile)

		if cmdOptions.Region != "" {
			setKey("region", cmdOptions.Region)
		}
	}

	err = config.SaveTo(cmdOptions.VaultConfigPath)
	if err != nil {
		log.Panic().Err(err).Str("file-path", cmdOptions.VaultConfigPath).Msg("could not save config")
	}
}
