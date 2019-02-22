/*
Copyright 2018 The Kubernetes Authors.

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
package manifest

import (
	"context"
	"flag"
	"fmt"

	"istio.io/istio/pkg/log"
)

var FlagChannel = "./channels"

func init() {
	// TODO: Yuk - global flags are ugly
	flag.StringVar(&FlagChannel, "channel", FlagChannel, "location of channel to use")
}

type ManifestController struct {
	repo Repository
}

func NewController() *ManifestController {
	// TODO: Accept as a parameter - but it's hard to have a flag per controller
	repo := NewFSRepository(FlagChannel)

	return &ManifestController{repo: repo}
}

func (c *ManifestController) ResolveManifest(ctx context.Context, packageName string, channelName string, version string) (string, error) {
	// TODO: We should actually do id (1.1.2-aws or 1.1.1-nginx). But maybe YAGNI
	id := version

	if id == "" {
		// TODO: Put channel in spec
		if channelName == "" {
			channelName = "stable"
		}

		channel, err := c.repo.LoadChannel(ctx, channelName)
		if err != nil {
			return "", err
		}

		version, err := channel.Latest()
		if err != nil {
			return "", err
		}

		// TODO: We should probably copy the kubelet componentconfig

		if version == nil {
			return "", fmt.Errorf("could not find latest version in channel %q", channelName)
		}
		id = version.Version

		log.Info("resolved version from channel")
	} else {
		log.Info("using specified version")
	}

	s, err := c.repo.LoadManifest(ctx, packageName, id)
	if err != nil {
		return "", fmt.Errorf("error loading manifest: %v", err)
	}

	return s, nil
}
