package util

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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/sts"
)

type PolicyDoc struct {
	Version   string
	Statement []map[string]interface{}
}

const (
	assumeAction = "sts:AssumeRole"
)

type AWSContext struct {
	org *organizations.Organizations
	iam *iam.IAM
	sts *sts.STS
}

func GetAWSContext() (client *AWSContext) {
	sess := session.Must(session.NewSession())

	config := aws.NewConfig()

	return &AWSContext{
		org: organizations.New(sess, config),
		iam: iam.New(sess, config),
		sts: sts.New(sess, config),
	}
}

func generateOrgRoleArns(accountMap map[string]string, role string) []string {
	var roles []string

	for accountId := range accountMap {
		roles = append(roles, fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, role))
	}

	return roles
}

func (ctx *AWSContext) GetRolesAndAccounts(role string) (roleArns []string, accountMap map[string]string) {
	cRoles := make(chan []string)
	cAccount := make(chan map[string]string)

	go func() {
		cRoles <- ctx.getRoles()
	}()

	go func() {
		cAccount <- ctx.getAccountNames()
	}()

	accountMap = <-cAccount
	close(cAccount)

	if role != "" {
		roleArns = generateOrgRoleArns(accountMap, role)
	}

	roleArns = append(roleArns, <-cRoles...)
	close(cRoles)

	return
}

type Profile struct {
	RoleArn     string
	RoleName    string
	ProfileName string
	AccountID   string
}

func GetProfiles(prefix string, accountMap map[string]string, roleArns []string, useRoleName bool) []Profile {
	var profiles []Profile

	for _, roleArn := range roleArns {
		// skip creating this profile if the role isn't a valid ARN (e.g. `*`)
		if !arn.IsARN(roleArn) {
			continue
		}

		role, _ := arn.Parse(roleArn)
		profileName, roleName := getProfileAndRoleName(accountMap, role, useRoleName)

		profiles = append(profiles, Profile{
			RoleArn:     roleArn,
			RoleName:    roleName,
			ProfileName: fmt.Sprint(prefix, profileName),
			AccountID:   role.AccountID,
		})
	}

	return profiles
}

func getProfileAndRoleName(accountMap map[string]string, role arn.ARN, useRoleName bool) (profileName string, roleName string) {
	if name, ok := accountMap[role.AccountID]; ok {
		profileName = name
	} else {
		profileName = role.AccountID
	}

	roleName = strings.Replace(role.Resource, "role/", "", 1)

	if useRoleName {
		profileName = fmt.Sprint(profileName, "_", roleName)
	}

	return
}

func getUser(userArn *string) *string {
	arnParts := strings.Split(*userArn, "/")
	return &arnParts[1]
}

func (ctx *AWSContext) getRoles() (roleArns []string) {
	log.Debug().Msg("getting caller identity")

	gcio, err := ctx.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Panic().Err(err).Msg("could not get caller identity")
	}

	log.Info().Str("user-arn", *gcio.Arn).Msg("Found user")

	lgfuo, err := ctx.iam.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: getUser(gcio.Arn),
	})
	if err != nil {
		log.Panic().Err(err).Str("user", *getUser(gcio.Arn)).Msg("could not list groups for user")
	}

	log.Debug().Msgf("Found %d groups", len(lgfuo.Groups))

	c := make(chan []string)

	for _, group := range lgfuo.Groups {
		go func(g iam.Group) {
			log.Debug().Str("group", *g.GroupName).Msg("Finding roles for group")
			c <- ctx.getRoleArnsForGroup(&g)
		}(*group)
	}

	for range lgfuo.Groups {
		roleArns = append(roleArns, (<-c)...)
	}

	log.Info().Msgf("Found %d roles", len(roleArns))
	log.Debug().Strs("roles", roleArns).Msgf("Roles")

	return
}

func (ctx *AWSContext) getAccountNames() map[string]string {
	accIDToName := map[string]string{}

	lai := &organizations.ListAccountsInput{}

	for {
		lao, err := ctx.org.ListAccounts(lai)
		if err != nil {
			log.Warn().Err(err).Msg("could not list organization member accounts")
			// ignore error so script can be used without these permissions
			break
		}

		log.Debug().Msgf("found %d member accounts", len(lao.Accounts))

		for _, acc := range lao.Accounts {
			accIDToName[*acc.Id] = *acc.Name
			log.Debug().
				Str("account-id", *acc.Id).
				Str("account-name", *acc.Name).
				Msg("found organization member account")
		}

		if lao.NextToken == nil {
			break
		}

		lai.NextToken = lao.NextToken
	}

	return accIDToName
}

func (ctx *AWSContext) getRoleArnsForGroup(group *iam.Group) (roles []string) {
	c := make(chan []string)

	go func() {
		c <- ctx.listInlinePolicyAndGetRoles(group)
	}()
	go func() {
		c <- ctx.listAttachedPolicyAndGetRoles(group)
	}()

	roles = append(roles, (<-c)...)
	roles = append(roles, (<-c)...)

	return
}

func (ctx *AWSContext) listInlinePolicyAndGetRoles(group *iam.Group) (roleArns []string) {
	log.Debug().Str("group", *group.GroupName).Msg("finding roles from group inline policies")

	lgpo, err := ctx.iam.ListGroupPolicies(&iam.ListGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		log.Panic().Err(err).Str("group", *group.GroupName).Msg("could not list inline group policies")
	}

	c := make(chan []string)

	for _, policy := range lgpo.PolicyNames {
		go func(p string) {
			log.Debug().Str("policy", p).Msg("Finding roles for inlined policy")
			c <- ctx.getRoleArnsForInlinePolicy(*group.GroupName, p)
		}(*policy)
	}

	for range lgpo.PolicyNames {
		roleArns = append(roleArns, (<-c)...)
	}

	return
}

func (ctx *AWSContext) getRoleArnsForInlinePolicy(group, policyName string) []string {
	ggpo, err := ctx.iam.GetGroupPolicy(&iam.GetGroupPolicyInput{
		GroupName:  &group,
		PolicyName: &policyName,
	})
	if err != nil {
		log.Panic().Err(err).Msg("could not get group policy")
	}

	return getRolesArnsFromPolicy(ggpo.PolicyDocument)
}

func (ctx *AWSContext) listAttachedPolicyAndGetRoles(group *iam.Group) (roleArns []string) {
	log.Debug().Str("group", *group.GroupName).Msg("finding roles from group attached policies")

	lagpo, err := ctx.iam.ListAttachedGroupPolicies(&iam.ListAttachedGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		log.Panic().Err(err).Str("group", *group.GroupName).Msg("could not list attached group policies")
	}

	c := make(chan []string)

	for _, policy := range lagpo.AttachedPolicies {
		go func(p iam.AttachedPolicy) {
			log.Debug().Str("policy ARN", *p.PolicyArn).Msg("Finding roles for attached policy")
			c <- ctx.getRoleArnsForAttachedPolicy(&p)
		}(*policy)
	}

	for range lagpo.AttachedPolicies {
		roleArns = append(roleArns, (<-c)...)
	}

	return
}

func (ctx *AWSContext) getRoleArnsForAttachedPolicy(policy *iam.AttachedPolicy) []string {
	gpio, err := ctx.iam.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: policy.PolicyArn,
	})
	if err != nil {
		log.Panic().Err(err).Str("policy-arn", *policy.PolicyArn).Msg("could not get policy")
	}

	gpvio, err := ctx.iam.GetPolicyVersion(&iam.GetPolicyVersionInput{
		PolicyArn: policy.PolicyArn,
		VersionId: gpio.Policy.DefaultVersionId,
	})
	if err != nil {
		log.Panic().Err(err).
			Str("policy", *policy.PolicyArn).Str("version-id", *gpio.Policy.DefaultVersionId).
			Msg("could not get policy version")
	}

	return getRolesArnsFromPolicy(gpvio.PolicyVersion.Document)
}

func getRolesArnsFromPolicy(policyJSON *string) (roles []string) {
	policyJson, err := url.QueryUnescape(*policyJSON)
	if err != nil {
		log.Panic().Err(err).Msg("could not unescape policy JSON")
	}

	var policyDoc PolicyDoc

	err = json.Unmarshal([]byte(policyJson), &policyDoc)
	if err != nil {
		log.Panic().Err(err).Msg("could not unmarshall policy JSON")
	}

	for _, statement := range policyDoc.Statement {
		if effStr, ok := statement["Effect"].(string); (ok && effStr != "Allow") || !checkAction(statement["Action"]) {
			continue
		}

		if resStr, ok := statement["Resource"].(string); ok {
			log.Debug().Str("role", resStr).Msg("found assumable role")
			roles = append(roles, resStr)
			continue
		}

		if resArr, ok := statement["Resource"].([]interface{}); ok {
			for _, res := range resArr {
				if resStr, ok := res.(string); ok {
					log.Debug().Str("role", resStr).Msg("found assumable role")
					roles = append(roles, resStr)
				}
			}
		}
	}

	return
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
