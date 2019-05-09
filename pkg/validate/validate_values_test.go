package validate

import (
	"fmt"
	"github.com/ostromart/istio-installer/pkg/util"
	"testing"

	"github.com/ghodss/yaml"
)

func TestValidateValues(t *testing.T) {
	tests := []struct {
		desc     string
		yamlStr  string
		wantErrs util.Errors
	}{
		{
			desc: "nil success",
		},
		{
			desc: "ProxyConfig",
			yamlStr: `
global:
  proxy:
    enabled: true
    includeIpRanges: "1.1.0.0/16,2.2.0.0/16"
    excludeIpRanges: "3.3.0.0/16,4.4.0.0/16"
    includeInboundPorts: "111,222"
    excludeInboundPorts: "333,444"
    clusterDomain: "my.domain"
    podDnsSearchNamespaces: "my-namespace"
    interceptionMode: TPROXY
    connectTimeout: "11s"
    drainDuration: "22s"
    parentShutdownDuration : "33s"
    concurrency: 5
`,
		},
		{
			desc: "BadIPRange",
			yamlStr: `
global:
  proxy:
    includeIpRanges: "1.1.0.256/16,2.2.0.257/16"
    excludeIpRanges: "3.3.0.0/33,4.4.0.0/34"
`,
			wantErrs: makeErrors([]string{`global.proxy.excludeIpRanges invalid CIDR address: 3.3.0.0/33`,
				`global.proxy.excludeIpRanges invalid CIDR address: 4.4.0.0/34`,
				`global.proxy.includeIpRanges invalid CIDR address: 1.1.0.256/16`,
				`global.proxy.includeIpRanges invalid CIDR address: 2.2.0.257/16`}),
		},
		{
			desc: "BadIPMalformed",
			yamlStr: `
global:
  proxy:
    includeIpRanges: "1.2.3/16,1.2.3.x/16"
`,
			wantErrs: makeErrors([]string{`global.proxy.includeIpRanges invalid CIDR address: 1.2.3/16`,
				`global.proxy.includeIpRanges invalid CIDR address: 1.2.3.x/16`}),
		},
		{
			desc: "BadPortRange",
			yamlStr: `
global:
  proxy:
    includeInboundPorts: "111,65536"
    excludeInboundPorts: "-1,444"
`,
			wantErrs: makeErrors([]string{`value global.proxy.excludeInboundPorts:-1 falls outside range [0, 65535]`,
				`value global.proxy.includeInboundPorts:65536 falls outside range [0, 65535]`}),
		},
		{
			desc: "BadPortMalformed",
			yamlStr: `
global:
  proxy:
    includeInboundPorts: "111,222x"
`,
			wantErrs: makeErrors([]string{`global.proxy.includeInboundPorts : strconv.ParseInt: parsing "222x": invalid syntax`}),
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			root := util.Tree{}
			err := yaml.Unmarshal([]byte(tt.yamlStr), &root)
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}
			errs := ValidateValues(root)
			if gotErr, wantErr := errs, tt.wantErrs; !util.EqualErrors(gotErr, wantErr) {
				t.Errorf("ValidateValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.yamlStr, gotErr, wantErr)
			}
		})
	}
}

func makeErrors(estr []string) util.Errors {
	var errs util.Errors
	for _, s := range estr {
		errs = util.AppendErr(errs, fmt.Errorf("%s", s))
	}
	return errs
}
