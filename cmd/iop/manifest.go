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

package iop

import (
	"fmt"
	"io/ioutil"

	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/validate"

	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/component/controlplane"
	"github.com/ostromart/istio-installer/pkg/util"
	"github.com/spf13/cobra"
)

func manifestCmd(rootArgs *rootArgs, printf, fatalf FormatFn) *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Generates Istio install manifest.",
		Long:  "The manifest subcommand is used to generate an Istio install manifest based on the input CR.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genManifest(rootArgs, printf, fatalf)
		}}

}

func genManifest(args *rootArgs, printf, fatalf FormatFn) {
	if args.crPath == "" {
		fatalf("Must set crpath")
	}
	b, err := ioutil.ReadFile(args.crPath)
	if err != nil {
		fatalf(err.Error())
	}
	icp := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(string(b), icp); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(icp); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	cp := controlplane.NewIstioControlPlane(icp, component.V12DirLayout)
	if err := cp.Run(); err != nil {
		fatalf(err.Error())
	}

	y, errs := cp.RenderManifest()
	err = errs.ToError()
	if err != nil {
		fatalf(err.Error())
	}
	fmt.Println(y)
}
