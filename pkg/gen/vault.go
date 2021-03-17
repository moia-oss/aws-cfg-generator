package gen

import (
	"fmt"

	"github.com/moia-oss/aws-cfg-generator/pkg/cmd"
	"github.com/moia-oss/aws-cfg-generator/pkg/util"

	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

func GenerateVaultProfile(accountMap map[string]string, roleArns []string, cmdOptions cmd.VaultCmd) {
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
		panic(err)
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
