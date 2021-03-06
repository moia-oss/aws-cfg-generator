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
	"github.com/rs/zerolog/log"
	"gopkg.in/ini.v1"
)

func GetKeySetter(section *ini.Section) func(key, value string) {
	return func(key, value string) {
		_, err := section.NewKey(key, value)
		if err != nil {
			log.Panic().Err(err).
				Str("section", section.Name()).Str("key", key).Str("val", value).
				Msg("could not set")
		}
	}
}
