package kubectlcmd

import (
	"context"
	"errors"
	"testing"

	"io/ioutil"
	"reflect"

	"os/exec"
)

// collector is a commandSite implementation that stubs cmd.Run() calls for tests
type collector struct {
	Error error
	Cmds  []*exec.Cmd
}

func (s *collector) Run(c *exec.Cmd) error {
	s.Cmds = append(s.Cmds, c)
	return s.Error
}

func TestKubectlApply(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		manifest   string
		args       []string
		err        error
		expectArgs []string
	}{
		{
			name:       "manifest",
			namespace:  "",
			manifest:   "foo",
			expectArgs: []string{"kubectl", "apply", "-f", "-"},
		},
		{
			name:       "manifest with apply",
			namespace:  "kube-system",
			manifest:   "heynow",
			expectArgs: []string{"kubectl", "apply", "-n", "kube-system", "-f", "-"},
		},
		{
			name:       "error propagation",
			expectArgs: []string{"kubectl", "apply", "-f", "-"},
			err:        errors.New("error"),
		},
		{
			name:       "manifest with prune",
			namespace:  "kube-system",
			manifest:   "heynow",
			args:       []string{"--prune=true", "--prune-whitelist=hello-world"},
			expectArgs: []string{"kubectl", "apply", "-n", "kube-system", "--prune=true", "--prune-whitelist=hello-world", "-f", "-"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cs := collector{Error: test.err}
			kubectl := &Client{cmdSite: &cs}
			err := kubectl.Apply(context.Background(), test.namespace, test.manifest, test.args...)

			if test.err != nil && err == nil {
				t.Error("expected error to occur")
			} else if test.err == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(cs.Cmds) != 1 {
				t.Errorf("expected 1 command to be invoked, got: %d", len(cs.Cmds))
			}

			cmd := cs.Cmds[0]
			if !reflect.DeepEqual(cmd.Args, test.expectArgs) {
				t.Errorf("argument mistmatch, expected: %v, got: %v", test.expectArgs, cmd.Args)
			}

			stdinBytes, err := ioutil.ReadAll(cmd.Stdin)
			if stdin := string(stdinBytes); stdin != test.manifest {
				t.Errorf("manifest mismatch, expected: %v, got: %v", test.manifest, stdin)
			}
		})
	}

}
