package cmd

// nolint:govet // we need the bare `required` tag here
type VaultCmd struct {
	SourceProfile        string `help:"The profile that your credentials should come from" default:"default"`
	Region               string `help:"Override the region configured with your source profile"`
	VaultConfigPath      string `help:"Where to load/save the config" required`
	KeepCustomConfig     bool   `help:"Retains any custom profiles or settings. Set to false to remove everything except the source profile and generated config" default:true`
	UseRoleNameInProfile bool   `help:"Append the role name to the profile name" default:false`
}
