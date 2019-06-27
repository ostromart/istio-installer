// Copyright 2019 Istio Authors
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

package iop

import (
	"flag"

	"github.com/spf13/cobra"

	"istio.io/pkg/version"
)

type rootArgs struct {
	// inFilename is the path to the input IstioIstall CR.
	inFilename string
	// outFilename is the path to the generated output filename.
	outFilename string
	// logToStdErr controls whether logs are sent to stderr.
	logToStdErr bool
	// Dry run performs all steps except actually applying the manifests or creating output dirs/files.
	dryRun bool
	// Verbose controls whether additional debug output is displayed and logged.
	verbose bool
}

func addFlags(cmd *cobra.Command, rootArgs *rootArgs) {
	cmd.PersistentFlags().StringVarP(&rootArgs.inFilename, "filename", "f", "",
		"The path to the input IstioIstall CR. Uses in cluster value with kubectl if unset.")
	cmd.PersistentFlags().StringVarP(&rootArgs.outFilename, "output", "o",
		"", "Manifest output path.")
	cmd.PersistentFlags().BoolVarP(&rootArgs.logToStdErr, "logtostderr", "",
		false, "Send logs to stderr.")
	cmd.PersistentFlags().BoolVarP(&rootArgs.dryRun, "dry-run", "",
		true, "Console/log output only, make no changes.")
	cmd.PersistentFlags().BoolVarP(&rootArgs.verbose, "verbose", "",
		false, "Verbose output.")
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "iop",
		Short: "Command line Istio install utility.",
		Long: "This command uses the Istio operator code to generate templates, query configurations and perform " +
			"utility operations.",
	}
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootArgs := &rootArgs{}

	ic := installCmd(rootArgs)
	mc := manifestCmd(rootArgs)
	dpc := dumpProfileDefaultsCmd(rootArgs)

	addFlags(ic, rootArgs)
	addFlags(mc, rootArgs)
	addFlags(dpc, rootArgs)

	rootCmd.AddCommand(ic)
	rootCmd.AddCommand(mc)
	rootCmd.AddCommand(dpc)
	rootCmd.AddCommand(version.CobraCommand())

	return rootCmd
}
