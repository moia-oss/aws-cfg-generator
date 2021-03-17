package cmd

import (
	"github.com/moia-oss/aws-cfg-generator/pkg/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

// nolint:govet // we need the bare `required` tag here
type SwitchRolesCmd struct {
	Color                string `help:"The hexcode color that should be set for each profile" default:"00ff7f"`
	OutputFile           string `help:"Where to save the config." required`
	UseRoleNameInProfile bool   `help:"Append the role name to the profile name" default:false`
}

func (swc *SwitchRolesCmd) Run(cli *CLI) error {
	roleArns, accountMap := util.GetAWSContext().GetRolesAndAccounts()
	generateSwitchRolesProfile(accountMap, roleArns, cli.SwitchRoles)

	return nil
}

func generateSwitchRolesProfile(accountMap map[string]string, roleArns []string, cmdOptions SwitchRolesCmd) {
	config := ini.Empty()

	for _, profile := range util.GetProfiles("", accountMap, roleArns, cmdOptions.UseRoleNameInProfile) {
		profileSection := config.Section(profile.ProfileName)

		setKey := util.GetKeySetter(profileSection)

		setKey("aws_account_id", profile.AccountID)
		setKey("role_name", profile.RoleName)
		setKey("color", cmdOptions.Color)
	}

	err := config.SaveTo(cmdOptions.OutputFile)
	if err != nil {
		log.Panic().Err(err).Str("file-path", cmdOptions.OutputFile).Msg("could not save file")
	}
}
