#!/bin/bash

set -eu

PERM_PACKAGE="code.cloudfoundry.org/perm"
# shellchheck disable=SC2153
PERM_PACKAGE_PATH="${GOPATH}/src/${PERM_PACKAGE}"
PROTOS_PATH="${PERM_PACKAGE_PATH}/protos"

function compile_go_protos() {
  go install "${PERM_PACKAGE}/vendor/github.com/gogo/protobuf/protoc-gen-gofast"

  goout="${PERM_PACKAGE_PATH}/internal/protos"
  mkdir -p "$goout"

  protoc \
    --gofast_out=plugins=grpc:"${goout}" \
    --plugin=protoc-gen-grpc \
    -I="${PROTOS_PATH}:${PERM_PACKAGE_PATH}/vendor" \
    "${PROTOS_PATH}/"*.proto
}

function compile_ruby_protos() {
  perm_rb_path="${HOME}/workspace/perm-rb"
  rubyout="${perm_rb_path}/lib/perm/protos"
  mkdir -p "$rubyout"

  set +e
  ruby_protoc_plugin="$(command -v grpc_tools_ruby_protoc_plugin)"
  set -e
  : "${ruby_protoc_plugin:?"Did not find grpc_tools_ruby_protoc_plugin"}"

  protoc \
    --ruby_out="$rubyout" \
    --plugin=protoc-gen-grpc="$ruby_protoc_plugin" \
    --grpc_out="$rubyout" \
    -I="${PROTOS_PATH}:${PERM_PACKAGE_PATH}/vendor" \
    "${PROTOS_PATH}/"*.proto
}

function main() {
  compile_go_protos
  compile_ruby_protos
}

main
