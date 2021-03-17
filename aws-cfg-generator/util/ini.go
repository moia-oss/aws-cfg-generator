package util

import "gopkg.in/ini.v1"

func GetKeySetter(section *ini.Section) func(key, value string) {
	return func(key, value string) {
		_, err := section.NewKey(key, value)
		if err != nil {
			panic(err)
		}
	}
}
