package util

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

func (ctx *AWSContext) GetRolesAndAccounts() (roleArns []string, accountMap map[string]string) {
	accountMap = ctx.getAccountNames()

	gcio, err := ctx.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		log.Panic().Err(err).Msg("coulld not get calleer identity")
	}

	log.Info().Str("user-arn", *gcio.Arn).Msg("Found user")

	lgfuo, err := ctx.iam.ListGroupsForUser(&iam.ListGroupsForUserInput{
		UserName: getUser(gcio.Arn),
	})
	if err != nil {
		log.Panic().Err(err).Str("user", *getUser(gcio.Arn)).Msg("could not list groups for user")
	}

	for _, group := range lgfuo.Groups {
		log.Debug().Str("group", *group.GroupName).Msg("Finding roles for group")
		roleArns = append(roleArns, ctx.getRoleArnsForGroup(group)...)
	}

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

		for _, acc := range lao.Accounts {
			accIDToName[*acc.Id] = *acc.Name
			log.Debug().Str("account-id", *acc.Id).Str("account-name", *acc.Name).
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
	lgpo, err := ctx.iam.ListGroupPolicies(&iam.ListGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		panic(err)
	}
	for _, policy := range lgpo.PolicyNames {
		log.Debug().Str("policy", *policy).Msg("Finding roles for inlined policy")
		roles = append(roles, ctx.getRoleArnsForInlinePolicy(group.GroupName, policy)...)
	}

	lagpo, err := ctx.iam.ListAttachedGroupPolicies(&iam.ListAttachedGroupPoliciesInput{
		GroupName: group.GroupName,
	})
	if err != nil {
		panic(err)
	}
	for _, policy := range lagpo.AttachedPolicies {
		log.Debug().Str("policy ARN", *policy.PolicyArn).Msg("Finding roles for attached policy")
		roles = append(roles, ctx.getRoleArnsForAttachedPolicy(policy)...)
	}

	return
}

func (ctx *AWSContext) getRoleArnsForInlinePolicy(group, policyName *string) []string {
	ggpo, err := ctx.iam.GetGroupPolicy(&iam.GetGroupPolicyInput{
		GroupName:  group,
		PolicyName: policyName,
	})
	if err != nil {
		log.Panic().Err(err).Msg("could not get group policy")
	}

	return getRolesArnsFromPolicy(ggpo.PolicyDocument)
}

func (ctx *AWSContext) getRoleArnsForAttachedPolicy(policy *iam.AttachedPolicy) []string {
	gpio, err := ctx.iam.GetPolicy(&iam.GetPolicyInput{
		PolicyArn: policy.PolicyArn,
	})
	if err != nil {
		panic(err)
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
