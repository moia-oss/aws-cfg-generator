package main

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

	"github.com/rs/zerolog"

	"github.com/alecthomas/kong"

	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/cmd"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/gen"
	"github.com/moia-oss/aws-cfg-generator/aws-cfg-generator/util"
)

// nolint:govet // we need the bare `cmd` tag here
type CLI struct {
	Vault       cmd.VaultCmd       `cmd help:"generates a config for aws-vault"`
	SwitchRoles cmd.SwitchRolesCmd `cmd help:"generates a config for aws-extend-switch-roles"`
	Debug       bool               `help:"set the log level to debug" default:false`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	if cli.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	switch ctx.Command() {
	case "vault":
		roleArns, accountMap := util.GetAWSContext().GetRolesAndAccounts()
		gen.GenerateVaultProfile(accountMap, roleArns, cli.Vault)

	case "switch-roles":
		roleArns, accountMap := util.GetAWSContext().GetRolesAndAccounts()
		gen.GenerateSwitchRolesProfile(accountMap, roleArns, cli.SwitchRoles)
	default:
		panic(fmt.Errorf("unsupported command '%s'", ctx.Command()))
	}
}
