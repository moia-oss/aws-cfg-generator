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
	"strings"

	"github.com/moia-oss/aws-cfg-generator/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

// nolint:govet // we need the bare `required` tag here
type SwitchRolesCmd struct {
	Color                string `help:"The hexcode color that should be set for each profile which name doesn't end in 'prd' or 'global'" default:"00ff7f"`
	DevColor             string `help:"The hexcode color that should be set for each profile which name ends in 'dev' or 'poc'" default:"00d619"`
	IntColor             string `help:"The hexcode color that should be set for each profile which name ends in 'int' or 'stg'" default:"ffea00"`
	PrdColor             string `help:"The hexcode color that should be set for each profile which name ends in 'prd' or 'global'" default:"ff0000"`
	OutputFile           string `help:"Where to save the config." required`
	UseRoleNameInProfile bool   `help:"Append the role name to the profile name" default:false`
}

func (swc *SwitchRolesCmd) Run(cli *CLI) error {
	roleArns, accountMap := util.GetAWSContext().GetRolesAndAccounts(cli.Role)
	generateSwitchRolesProfile(accountMap, roleArns, cli.SwitchRoles, cli.Ordered)

	return nil
}

func envSpecificColor(profileName string, cmdOptions SwitchRolesCmd) string {
	lowerKeyProfileName := strings.ToLower(profileName)

	if strings.HasSuffix(lowerKeyProfileName, "dev") || strings.HasSuffix(lowerKeyProfileName, "poc") {
		return cmdOptions.DevColor
	}

	if strings.HasSuffix(lowerKeyProfileName, "int") || strings.HasSuffix(lowerKeyProfileName, "stg") {
		return cmdOptions.IntColor
	}

	if strings.HasSuffix(lowerKeyProfileName, "prd") || strings.HasSuffix(lowerKeyProfileName, "global") {
		return cmdOptions.PrdColor
	}

	return cmdOptions.Color
}

func generateSwitchRolesProfile(accountMap map[string]string, roleArns []string, cmdOptions SwitchRolesCmd, ordered bool) {
	config := ini.Empty()

	profiles := util.GetProfiles("", accountMap, roleArns, cmdOptions.UseRoleNameInProfile)

	if ordered {
		profiles = util.OrderProfiles(profiles)
	}

	for _, profile := range profiles {
		profileSection := config.Section(profile.ProfileName)

		setKey := util.GetKeySetter(profileSection)

		setKey("aws_account_id", profile.AccountID)
		setKey("role_name", profile.RoleName)
		setKey("color", envSpecificColor(profile.ProfileName, cmdOptions))
	}

	err := config.SaveTo(cmdOptions.OutputFile)
	if err != nil {
		log.Panic().Err(err).Str("file-path", cmdOptions.OutputFile).Msg("could not save file")
	}
}
