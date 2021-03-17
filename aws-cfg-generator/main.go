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
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/moia-oss/aws-cfg-generator/pkg/cmd"
)

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli)

	if cli.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	err := ctx.Run(cli)
	if err != nil {
		log.Panic().Err(err).Msgf("unexpected CLI error")
	}
}
