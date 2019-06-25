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
	"os"

	"github.com/spf13/cobra"
	"istio.io/pkg/log"

	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/component/controlplane"
	"github.com/ostromart/istio-installer/pkg/helm"
	"github.com/ostromart/istio-installer/pkg/translate"
	"github.com/ostromart/istio-installer/pkg/util"
	"github.com/ostromart/istio-installer/pkg/validate"
	"github.com/ostromart/istio-installer/pkg/version"
)

func manifestCmd(rootArgs *rootArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Generates Istio install manifest.",
		Long:  "The manifest subcommand is used to generate an Istio install manifest based on the input CR.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genManifest(rootArgs)
		}}

}

func genManifest(args *rootArgs) {
	if err := configLogs(args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not configure logs: %s", err)
		os.Exit(1)
	}

	writer, err := getWriter(args)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Errorf("Did not close output successfully: %v", err)
		}
	}()

	overlayYAML := ""
	if args.inFilename != "" {
		b, err := ioutil.ReadFile(args.inFilename)
		if err != nil {
			log.Fatalf("Could not open input file: %s", err)
		}
		overlayYAML = string(b)
	}

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		log.Fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS); len(errs) != 0 {
		log.Fatalf(errs.ToError().Error())
	}

	// Now read the base profile specified in the user spec. If nothing specified, use default.
	baseYAML, err := helm.ReadValuesYAML(overlayICPS.BaseProfilePath)
	if err != nil {
		log.Fatalf(err.Error())
	}
	// Unmarshal and validate the base CR.
	baseICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(baseYAML, baseICPS); err != nil {
		log.Fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(baseICPS); len(errs) != 0 {
		log.Fatalf(errs.ToError().Error())
	}

	mergedYAML, err := helm.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Now unmarshal and validate the combined base profile and user CR overlay.
	mergedcps := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(mergedYAML, mergedcps); err != nil {
		log.Fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(mergedcps); len(errs) != 0 {
		log.Fatalf(errs.ToError().Error())
	}

	if yd := util.YAMLDiff(mergedYAML, util.ToYAMLWithJSONPB(mergedcps)); yd != "" {
		log.Fatalf("Validated YAML differs from input: \n%s", yd)
	}

	// TODO: remove version hard coding.
	cp := controlplane.NewIstioControlPlane(mergedcps, translate.Translators[version.MinorVersion{Major: 1, Minor: 2}])
	if err := cp.Run(); err != nil {
		log.Fatalf(err.Error())
	}

	y, errs := cp.RenderManifest()
	err = errs.ToError()
	if err != nil {
		log.Fatalf(err.Error())
	}
	writer.WriteString(y)
}
