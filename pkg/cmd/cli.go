package cmd

// nolint:govet // we need the bare `cmd` tag here
type CLI struct {
	Vault       VaultCmd       `cmd help:"generates a config for aws-vault"`
	SwitchRoles SwitchRolesCmd `cmd help:"generates a config for aws-extend-switch-roles"`
	Debug       bool           `help:"set the log level to debug" default:false`
}
