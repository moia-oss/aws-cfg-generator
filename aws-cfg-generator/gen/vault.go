package gen

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/cmd"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/util"
	"gopkg.in/ini.v1"
)

func GenerateVaultProfile(accountMap map[string]string, roleArns []string, cmdOptions cmd.VaultCmd) {
	config, err := ini.Load(cmdOptions.VaultConfigPath)

	if err != nil {
		panic(err)
	}

	sourceProfileSectionName := cmdOptions.SourceProfile

	// the profile can either be [default] or [profile foo]
	if sourceProfileSectionName != "default" {
		sourceProfileSectionName = fmt.Sprint("profile ", sourceProfileSectionName)
	}

	// make sure the source section exists
	sourceProfileSection, err := config.GetSection(sourceProfileSectionName)

	if err != nil {
		panic(err)
	}

	// only copy the source profile, discard the rest of the config
	if !cmdOptions.KeepCustomConfig {
		config = ini.Empty()

		newSourceProfileSection, err := config.NewSection(sourceProfileSectionName)

		if err != nil {
			panic(err)
		}

		setKey := util.GetKeySetter(newSourceProfileSection)

		for key, value := range sourceProfileSection.KeysHash() {
			setKey(key, value)
		}
	}

	for _, roleArn := range roleArns {
		// skip creating this profile if the role isn't a valid ARN (e.g. `*`)
		if !arn.IsARN(roleArn) {
			return
		}

		role, _ := arn.Parse(roleArn)

		profileName, _ := util.GetProfileAndRoleName(accountMap, role, cmdOptions.UseRoleNameInProfile)

		sectionName := fmt.Sprint("profile ", profileName)

		profileSection := config.Section(sectionName)

		setKey := util.GetKeySetter(profileSection)

		setKey("role_arn", roleArn)
		setKey("source_profile", cmdOptions.SourceProfile)
		setKey("include_profile", cmdOptions.SourceProfile)

		if cmdOptions.Region != "" {
			setKey("region", cmdOptions.Region)
		}
	}

	err = config.SaveTo(cmdOptions.VaultConfigPath)

	if err != nil {
		panic(err)
	}
}
