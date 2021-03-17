package gen

import (
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/cmd"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/util"

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
		panic(err)
	}
}
