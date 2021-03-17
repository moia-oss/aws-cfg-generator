package gen

import (
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/cmd"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/util"

	"github.com/aws/aws-sdk-go/aws/arn"
	"gopkg.in/ini.v1"
)

func GenerateSwitchRolesProfile(accountMap map[string]string, roleArns []string, cmdOptions cmd.SwitchRolesCmd) {
	config := ini.Empty()

	for _, roleArn := range roleArns {
		if !arn.IsARN(roleArn) {
			return
		}

		role, _ := arn.Parse(roleArn)

		profileName, roleName := util.GetProfileAndRoleName(accountMap, role, cmdOptions.UseRoleNameInProfile)

		profileSection := config.Section(profileName)

		setKey := util.GetKeySetter(profileSection)

		setKey("aws_account_id", role.AccountID)
		setKey("role_name", roleName)
		setKey("color", cmdOptions.Color)
	}

	err := config.SaveTo(cmdOptions.OutputFile)

	if err != nil {
		panic(err)
	}
}
