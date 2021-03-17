package cmd

// nolint:govet // we need the bare `required` tag here
type SwitchRolesCmd struct {
	Color                string `help:"The hexcode color that should be set for each profile" default:"00ff7f"`
	OutputFile           string `help:"Where to save the config." required`
	UseRoleNameInProfile bool   `help:"Append the role name to the profile name" default:false`
}
