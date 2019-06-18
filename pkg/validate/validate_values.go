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

package validate

import (
	"github.com/ostromart/istio-installer/pkg/util"
)

var (
	// defaultValidations maps a data path to a validation function.
	defaultValuesValidations = map[string]ValidatorFunc{
		"global.proxy.includeIpRanges":     validateStringList(validateCIDR),
		"global.proxy.excludeIpRanges":     validateStringList(validateCIDR),
		"global.proxy.includeInboundPorts": validateStringList(validatePortNumberString),
		"global.proxy.excludeInboundPorts": validateStringList(validatePortNumberString),
	}

	// requiredValues lists all the values that must be non-empty.
	requiredSetValues = map[string]bool{}
)

// CheckValues validates the values in the given tree, which follows the Istio values.yaml schema.
func CheckValues(root util.Tree) util.Errors {
	return validateValues(defaultValuesValidations, root, nil)
}

func validateValues(validations map[string]ValidatorFunc, node interface{}, path util.Path) (errs util.Errors) {
	pstr := path.String()
	dbgPrint("validateValues %s", pstr)
	vf := defaultValuesValidations[pstr]
	if vf != nil {
		errs = util.AppendErrs(errs, vf(path, node))
	}

	nn, ok := node.(util.Tree)
	if !ok {
		nn, ok = node.(map[string]interface{})
		if !ok {
			// Leaf, nothing more to recurse.
			return
		}
	}
	for k, v := range nn {
		errs = util.AppendErrs(errs, validateValues(validations, v, append(path, k)))
	}

	return errs
}
