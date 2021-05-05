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

// nolint:govet // we need the bare `cmd` tag here
type CLI struct {
	Vault       VaultCmd       `cmd help:"generates a config for aws-vault"`
	SwitchRoles SwitchRolesCmd `cmd help:"generates a config for aws-extend-switch-roles"`
	Debug       bool           `help:"set the log level to debug" default:"false"`
	Role        string         `help:"If set, then a profile with this role will be generated for every account in the organization, in addition to the roles that the user has permissions to assume"`
}
