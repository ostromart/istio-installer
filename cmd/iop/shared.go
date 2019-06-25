// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package iop contains types and functions that are used across the full
// set of mixer commands.
package iop

import (
	"os"

	"istio.io/pkg/log"
)

const (
	logFilePath = "./iop.log"
)

func getWriter(args *rootArgs) (*os.File, error) {
	writer := os.Stdout
	if args.outFilename != "" {
		file, err := os.Create(args.outFilename)
		if err != nil {
			log.Fatalf("Could not open output file: %s", err)
		}

		writer = file
	}
	return writer, nil
}

func configLogs(args *rootArgs) error {
	opt := log.DefaultOptions()
	if !args.logToStdErr {
		opt.ErrorOutputPaths = []string{logFilePath}
		opt.OutputPaths = []string{logFilePath}
	}
	return log.Configure(opt)
}
