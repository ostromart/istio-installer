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
	"fmt"
	"io/ioutil"

	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/util"
	"github.com/ostromart/istio-installer/pkg/validate"
	"github.com/spf13/cobra"
)

func dumpProfileDefaultsCmd(rootArgs *rootArgs, printf, fatalf FormatFn) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump-profile-defaults",
		Short: "Dump default values for the profile passed in the CR.",
		Long:  "The dump-defaults subcommand is used to dump default values for the profile passed in the CR.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			dumpProfile(rootArgs, printf, fatalf)
		}}
	return cmd
}

func dumpProfile(args *rootArgs, printf, fatalf FormatFn) {
	if args.crPath == "" {
		fatalf("Must set crpath")
	}
	b, err := ioutil.ReadFile(args.crPath)
	if err != nil {
		fatalf(err.Error())
	}
	overlayYAML := string(b)

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	// Now read the base profile specified in the user spec.
	b, err = ioutil.ReadFile(util.GetLocalFilePath(overlayICPS.BaseProfilePath))
	if err != nil {
		fmt.Printf("1")
		fatalf(err.Error())
	}
	fmt.Println(string(b))
}
