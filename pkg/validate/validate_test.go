package validate

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		desc    string
		yamlStr string
		wantErr string
	}{
		{
			desc: "nil success",
		},
		{
			desc: "PilotConfig",
			yamlStr: `
trafficManagement:
  pilot:
    sidecar: true
`,
		},
		{
			desc: "ProxyConfig",
			yamlStr: `
trafficManagement:
  enabled: true
  includeIpRanges: "1.1.0.0/16,2.2.0.0/16"
  excludeIpRanges: "3.3.0.0/16,4.4.0.0/16"
  includeInboundPorts: "111,222"
  excludeInboundPorts: "333,444"
  clusterDomain: "my.domain"
  podDnsSearchNamespaces: "my-namespace"
  enableAutoInjection: true
  enableNamespacesByDefault: true
  proxy:
    interceptionMode: TPROXY
    connectTimeout: "11s"
    drainDuration: "22s"
    parentShutdownDuration : "33s"
    concurrency: 5
`,
		},
		{
			desc: "CommonConfig",
			yamlStr: `
trafficManagement:
  proxy:
    common:
      enabled: true
      namespace: istio-control-system
      debug: INFO
      env:
        aa: bb
        cc: dd
      args:
        b: b
        c: d
      k8s:
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        readinessProbe:
          initialDelaySeconds: 11
          periodSeconds: 22
          successThreshold: 33
          failureThreshold: 44
        hpaSpec:
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            name: php-apache
          minReplicas: 1
          maxReplicas: 10
          targetCPUUtilizationPercentage: 80
        nodeSelector:
          disktype: ssd
`,
		},
		{
			desc: "BadIPRange",
			yamlStr: `
trafficManagement:
  includeIpRanges: "1.1.0.256/16,2.2.0.257/16"
  excludeIpRanges: "3.3.0.0/33,4.4.0.0/34"
`,
			wantErr: "TrafficManagement/IncludeIpRanges invalid CIDR address: 1.1.0.256/16, " +
				"TrafficManagement/IncludeIpRanges invalid CIDR address: 2.2.0.257/16, " +
				"TrafficManagement/ExcludeIpRanges invalid CIDR address: 3.3.0.0/33, "+
				"TrafficManagement/ExcludeIpRanges invalid CIDR address: 4.4.0.0/34",
		},
		{
			desc: "BadIPMalformed",
			yamlStr: `
trafficManagement:
  includeIpRanges: "1.2.3/16,1.2.3.x/16"
`,
			wantErr: "TrafficManagement/IncludeIpRanges invalid CIDR address: 1.2.3/16, " +
				"TrafficManagement/IncludeIpRanges invalid CIDR address: 1.2.3.x/16",
		},
		{
			desc: "BadPortRange",
			yamlStr: `
trafficManagement:
  includeInboundPorts: "111,65536"
  excludeInboundPorts: "-1,444"
`,
			wantErr: "value TrafficManagement/IncludeInboundPorts:65536 falls outside range [0, 65535], "+
				"value TrafficManagement/ExcludeInboundPorts:-1 falls outside range [0, 65535]",
		},
		{
			desc: "BadPortMalformed",
			yamlStr: `
trafficManagement:
  includeInboundPorts: "111,222x"
`,
			wantErr: `TrafficManagement/IncludeInboundPorts : strconv.ParseInt: parsing "222x": invalid syntax`,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			ispec := &v1alpha1.InstallerSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			err = Validate(defaultValidations, ispec)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ProtoToValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.yamlStr, gotErr, wantErr)
			}
		})
	}
}

func unmarshalWithJSONPB(y string, out proto.Message) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	u := jsonpb.Unmarshaler{}
	err = u.Unmarshal(bytes.NewReader(jb), out)
	if err != nil {
		return err
	}
	return nil
}

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
