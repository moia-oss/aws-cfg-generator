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
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
)

const (
	assumeAction = "sts:AssumeRole"
)

type VaultModel struct {
	ProfileName   string
	Region        string
	SourceProfile string
	RoleArn       string
}

const vaultTemplate = `[profile {{.ProfileName}}]
{{if (not (eq .Region "")) }}region={{.Region}}
{{end}}role_arn={{.RoleArn}}
source_profile={{.SourceProfile}}
input_profile={{.SourceProfile}}

`

var switchRolesTemplate = `[%s]
aws_account_id = %s
role_name = %s
color = %s

`

// nolint:govet // we need the bare `cmd` tag here
type CLI struct {
	Vault       VaultCmd       `cmd help:"generates a config for aws-vault"`
	SwitchRoles SwitchRolesCmd `cmd help:"generates a config for aws-extend-switch-roles"`
}

type VaultCmd struct {
	SourceProfile string `help:"The profile that your credentials should come from" default:"default"`
	Region        string `help:"Override the region configured with your source profile"`
}

type SwitchRolesCmd struct {
	Color string `help:"The hexcode color that should be set for each profile" default:"00ff7f"`
}

type configCreator struct {
	iamClient       *iam.IAM
	generateProfile func(role string)
}

type PolicyDoc struct {
	Version   string
	Statement []map[string]interface{}
}

func getUser(userArn *string) *string {
	arnParts := strings.Split(*userArn, "/")
	return &arnParts[1]
}

func generateSwitchRolesProfile(role, color string) {
	if !arn.IsARN(role) {
		return
	}

	roleArn, _ := arn.Parse(role)

	roleSplit := strings.Split(roleArn.Resource, "/")

	fmt.Printf(switchRolesTemplate, roleArn.AccountID, roleArn.AccountID, roleSplit[1], color)
}

func generateVaultProfile(role, region, sourceProfile string) {
	// skip creating this profile if the role isn't a valid ARN (e.g. `*`)
	if !arn.IsARN(role) {
		return
	}

	roleArn, _ := arn.Parse(role)

	t := template.Must(template.New("vaultText").Parse(vaultTemplate))

	var b bytes.Buffer
	err := t.Execute(&b, VaultModel{
		ProfileName:   roleArn.AccountID,
		Region:        region,
		SourceProfile: sourceProfile,
		RoleArn:       role,
	})
	if err != nil {
		panic(err)
	}

	fmt.Print(b.String())
}

func checkAction(action interface{}) bool {
	if actionStr, ok := action.(string); ok && actionStr == assumeAction {
		return true
	}

	if actionArr, ok := action.([]interface{}); ok {
		for _, a := range actionArr {
			if aStr, ok := a.(string); ok && aStr == assumeAction {
				return true
			}
		}
	}

	return false
}

func getPolicyDocument(policyJSON *string) PolicyDoc {
	policyJson, err := url.QueryUnescape(*policyJSON)
	if err != nil {
		panic(err)
	}

	var policyDoc PolicyDoc
	err = json.Unmarshal([]byte(policyJson), &policyDoc)
	if err != nil {
		panic(err)
	}

	return policyDoc
}

func (cc *configCreator) statementHelper(polDoc PolicyDoc) {
	for _, statement := range polDoc.Statement {
		if effStr, ok := statement["Effect"].(string); (ok && effStr != "Allow") || !checkAction(statement["Action"]) {
			continue
		}

		if resStr, ok := statement["Resource"].(string); ok {
			cc.generateProfile(resStr)
			continue
		}

		if resArr, ok := statement["Resource"].([]interface{}); ok {
			for _, res := range resArr {
				if resStr, ok := res.(string); ok {
					cc.generateProfile(resStr)
				}
			}
		}
	}
}

func (cc *configCreator) generateCfgForInlinePolicy(group, policyName *string) {
	ggpo, err := cc.iamClient.GetGroupPolicy(&iam.GetGroupPolicyInput{
		GroupName:  group,
		PolicyName: policyName,
	})
	if err != nil {
		panic(err)
	}

	policyDoc := getPolicyDocument(ggpo.PolicyDocument)
	cc.statementHelper(policyDoc)
}

func (cc *configCreator) generateCfgsForAttachedPolicy(policy *iam.AttachedPolicy) {
	gpio, err := cc.iamClient.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: policy.PolicyArn,
	})
	if err != nil {
		panic(err)
	}

	gpvio, err := cc.iamClient.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: policy.PolicyArn,
		VersionId: gpio.Policy.DefaultVersionId,
	})
	if err != nil {
		panic(err)
	}

	policyDoc := getPolicyDocument(gpvio.PolicyVersion.Document)
	cc.statementHelper(policyDoc)
}

func (cc *configCreator) generateCfgsForGroup(group *iam.Group) {
	lgpo, err := cc.iamClient.ListGroupPolicies(&iam.ListGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		panic(err)
	}

	for _, policy := range lgpo.PolicyNames {
		cc.generateCfgForInlinePolicy(group.GroupName, policy)
	}

	lagpo, err := cc.iamClient.ListAttachedGroupPolicies(&iam.ListAttachedGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		panic(err)
	}
	for _, policy := range lagpo.AttachedPolicies {
		cc.generateCfgsForAttachedPolicy(policy)
	}
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli)

	sess := session.Must(session.NewSession())
	stsClient := sts.New(sess)

	gcio, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		panic(err)
	}

	user := getUser(gcio.Arn)
	var generatorFunc func(role string)

	switch ctx.Command() {
	case "vault":
		generatorFunc = func(role string) {
			generateVaultProfile(role, cli.Vault.Region, cli.Vault.SourceProfile)
		}
	case "switch-roles":
		generatorFunc = func(role string) {
			generateSwitchRolesProfile(role, cli.SwitchRoles.Color)
		}
	default:
		panic(fmt.Errorf("unsupported command '%s'", ctx.Command()))
	}

	iamClient := iam.New(sess)

	cfgCreator := &configCreator{
		iamClient:       iamClient,
		generateProfile: generatorFunc,
	}

	lgfuo, err := iamClient.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: user,
	})
	if err != nil {
		panic(err)
	}

	for _, group := range lgfuo.Groups {
		cfgCreator.generateCfgsForGroup(group)
	}
}
