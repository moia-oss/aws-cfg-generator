package gen

import (
	"github.com/moia-oss/aws-cfg-generator/pkg/cmd"
	"github.com/moia-oss/aws-cfg-generator/pkg/util"
	"github.com/rs/zerolog/log"

	"gopkg.in/ini.v1"
)

func GenerateSwitchRolesProfile(accountMap map[string]string, roleArns []string, cmdOptions cmd.SwitchRolesCmd) {
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
