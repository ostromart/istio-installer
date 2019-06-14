
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/ostromart/istio-installer/pkg/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

proto:
	protoc -I./vendor -I./vendor/github.com/gogo/protobuf/protobuf -I./pkg/apis/istio/v1alpha2/ --proto_path=pkg/apis/istio/v1alpha2/ --gofast_out=pkg/apis/istio/v1alpha2/ pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto
	sed -i -e 's|github.com/gogo/protobuf/protobuf/google/protobuf|github.com/gogo/protobuf/types|g' pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go
	go run ~/go/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go -O zz_generated.deepcopy -i ./pkg/apis/istio/v1alpha2/... -i ./vendor/github.com/gogo/protobuf/types/...
	patch pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go < pkg/apis/istio/v1alpha2/fixup_go_structs.patch

# Note: must add // +k8s:deepcopy-gen=package to doc.go in ./vendor/github.com/gogo/protobuf/types/ for types package
proto_gogo:
	go run ~/go/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go -v 5 -O zz_generated.deepcopy -i ./vendor/github.com/gogo/protobuf/types/...
	patch vendor/github.com/gogo/protobuf/types/zz_generated.deepcopy.go < vendor/github.com/gogo/protobuf/types/fixup_go_structs.patch

gen_patch:
	diff -u pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go.orig pkg/apis/istio/v1alpha2/istiocontrolplane_types.pb.go > pkg/apis/istio/v1alpha2/fixup_go_structs.patch || true

gen_gogo_patch:
	diff -u vendor/github.com/gogo/protobuf/types/zz_generated.deepcopy.orig.go vendor/github.com/gogo/protobuf/types/zz_generated.deepcopy.go > vendor/github.com/gogo/protobuf/types/fixup_go_structs.patch || true
