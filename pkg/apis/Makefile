all: generate

########################
# docker_gen
########################

# Use a different generation mechanism when running from the
# image itself
ifdef CIRCLECI
repo_dir = .
docker_gen = /usr/bin/protoc -I/protobuf -I$(repo_dir)
out_path = $(OUT_PATH)
docker_lock = protolock
docker_tool = prototool
else
#gen_img := gcr.io/istio-operator/protoc:2019-04-10
#lock_img := gcr.io/istio-operator/protolock:2019-04-10
#all_img := gcr.io/istio-operator/api-build-tools:2019-04-10
gen_img := gcr.io/istio-testing/protoc:2019-02-25
lock_img := gcr.io/istio-testing/protolock:2018-10-23
all_img := gcr.io/istio-testing/api-build-tools:2019-02-25
pwd := $(shell pwd)
mount_dir := /src
repo_dir := github.com/ostromart/istio-installer
repo_mount := $(mount_dir)/github.com/ostromart/istio-installer
docker_gen := docker run --rm -v $(pwd):$(repo_mount) -w $(mount_dir) $(gen_img) -I$(repo_dir)
out_path = .
docker_lock = docker run --rm -v $(pwd):$(repo_mount) -w $(repo_mount) $(lock_img)
docker_tool = docker run --rm -v $(pwd):$(repo_mount) -w $(repo_mount) $(all_img) prototool
endif

########################
# protoc_gen_gogo*
########################

gogo_plugin_prefix := --gogo_out=plugins=grpc,
gogofast_plugin_prefix := --gogofast_out=plugins=grpc,
gogoslick_plugin_prefix := --gogoslick_out=plugins=grpc,

########################
# protoc_gen_python
########################

protoc_gen_python_prefix := --python_out=,
protoc_gen_python_plugin := $(protoc_gen_python_prefix):$(repo_dir)/python/istio_api

comma := ,
empty:=
space := $(empty) $(empty)

importmaps := \
	gogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto \
	google/protobuf/any.proto=github.com/gogo/protobuf/types \
	google/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor \
	google/protobuf/duration.proto=github.com/gogo/protobuf/types \
	google/protobuf/struct.proto=github.com/gogo/protobuf/types \
	google/protobuf/timestamp.proto=github.com/gogo/protobuf/types \
	google/protobuf/wrappers.proto=github.com/gogo/protobuf/types \
	google/rpc/status.proto=github.com/gogo/googleapis/google/rpc \
	google/rpc/code.proto=github.com/gogo/googleapis/google/rpc \
	google/rpc/error_details.proto=github.com/gogo/googleapis/google/rpc \

# generate mapping directive with M<proto>:<go pkg>, format for each proto file
mapping_with_spaces := $(foreach map,$(importmaps),M$(map),)
gogo_mapping := $(subst $(space),$(empty),$(mapping_with_spaces))

gogo_plugin := $(gogo_plugin_prefix)$(gogo_mapping):$(out_path)
gogofast_plugin := $(gogofast_plugin_prefix)$(gogo_mapping):$(out_path)
gogoslick_plugin := $(gogoslick_plugin_prefix)$(gogo_mapping):$(out_path)

########################
# protoc_gen_docs
########################

protoc_gen_docs_plugin := --docs_out=warnings=true,dictionary=$(repo_dir)/dictionaries/en-US,custom_word_list=$(repo_dir)/dictionaries/custom.txt,mode=html_fragment_with_front_matter:$(repo_dir)/
protoc_gen_docs_plugin_for_networking := --docs_out=warnings=true,dictionary=$(repo_dir)/dictionaries/en-US,custom_word_list=$(repo_dir)/dictionaries/custom.txt,per_file=true,mode=html_fragment_with_front_matter:$(repo_dir)/

#####################
# Generation Rules
#####################

generate: \
	generate-operator-go 

#####################
# operator
#####################

config_operator_path := installer/v1alpha1
config_operator_protos := $(shell find $(config_operator_path) -type f -name '*.proto' | sort)
config_operator_pb_gos := $(config_operator_protos:.proto=.pb.go)
config_operator_pb_doc := $(config_operator_path)/istio.operator.v1alpha1.pb.html

generate-operator-go: $(config_operator_pb_gos) $(config_operator_pb_doc)

$(config_operator_pb_gos) $(config_operator_pb_doc): $(config_operator_protos)
	@$(docker_lock) status
	@$(docker_gen) $(gogofast_plugin) $(protoc_gen_docs_plugin)$(config_operator_path) $^

clean-operator:
	rm -f $(config_operator_pb_gos)
	rm -f $(config_operator_pb_doc)

#####################
# Lint
#####################

lint:
	@$(docker_tool) lint --protoc-bin-path=/usr/bin/protoc --protoc-wkt-path=/protobuf

#####################
# Cleanup
#####################

clean: 	clean-operator
