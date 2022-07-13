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
	"golang.org/x/exp/slices"
	"strings"
)

var stages = []string{"poc", "stg", "dev", "int", "prd"}

func profileLess(x, y Profile) bool {
	return x.ProfileName < y.ProfileName
}

func OrderProfiles(unorderedProfiles []Profile) []Profile {
	var singleProfiles, stagedProfiles []Profile

	for _, profile := range unorderedProfiles {
		if isStageProfile(profile) {
			stagedProfiles = append(stagedProfiles, profile)
		} else {
			singleProfiles = append(singleProfiles, profile)
		}

	}

	slices.SortFunc(singleProfiles, profileLess)
	slices.SortFunc(stagedProfiles, profileLess)

	return append(singleProfiles, stagedProfiles...)
}

func isStageProfile(profile Profile) bool {
	for _, stage := range stages {
		if strings.HasSuffix(profile.ProfileName, stage) {
			return true
		}
	}
	return false
}
