#!/bin/bash
set -exuo pipefail
IFS=$'\n\t'

tar -C "${GOPATH}/src/github.com/pachyderm/pachyderm" -xf /dev/stdin

# Make sure the proto compiler is reasonably up-to-date so that our compiled
# protobufs aren't generated by stale tools
max_age="$(date --date="1 month ago" +%s)"
if [[ "${max_age}" -gt "$(cat /last_run_time)" ]]; then
    PRE="\e[1;31m~\e[0m"
    echo -e "\e[1;31m~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\e[0m" >/dev/stderr
    echo -e "${PRE} \e[1;31mWARNING:\e[0m pachyderm_proto is out of date" >/dev/stderr
    echo -e "${PRE} please run" >/dev/stderr
    echo -e "${PRE}   make DOCKER_BUILD_FLAGS=--no-cache proto" >/dev/stderr
    echo -e "\e[1;31m~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\e[0m" >/dev/stderr
    exit 1
fi

cd "${GOPATH}/src/github.com/pachyderm/pachyderm"
mkdir -p v2/src
mkdir -p v2/src/internal/jsonschema

mapfile -t PROTOS < <(find src -name "*.proto" | sort)

for i in "${PROTOS[@]}"; do \
    if ! grep -q 'go_package' "${i}"; then
        echo -e "\e[1;31mError:\e[0m missing \"go_package\" declaration in ${i}" >/dev/stderr
        exit 1
    fi
done

protoc \
    -Isrc \
    --plugin=protoc-gen-zap="${GOPATH}/bin/protoc-gen-zap" \
    --plugin=protoc-gen-pach="${GOPATH}/bin/protoc-gen-pach" \
    --plugin=protoc-gen-doc="${GOPATH}/bin/protoc-gen-doc" \
    --plugin=protoc-gen-doc2="${GOPATH}/bin/protoc-gen-doc" \
    --plugin="${GOPATH}/bin/protoc-gen-jsonschema" \
    --plugin="${GOPATH}/bin/protoc-gen-validate" \
    --zap_out=":${GOPATH}/src" \
    --pach_out="v2/src" \
    --go_out=":${GOPATH}/src" \
    --go-grpc_out=":${GOPATH}/src" \
    --jsonschema_out="${GOPATH}/src/github.com/pachyderm/pachyderm/v2/src/internal/jsonschema" \
    --validate_out="lang=go,paths=:${GOPATH}/src" \
    --doc_out="${GOPATH}/src/github.com/pachyderm/pachyderm/v2" \
    --doc2_out="${GOPATH}/src/github.com/pachyderm/pachyderm/v2" \
    --jsonschema_opt="enforce_oneof" \
    --jsonschema_opt="file_extension=schema.json" \
    --jsonschema_opt="disallow_additional_properties" \
    --jsonschema_opt="enums_as_strings_only" \
    --jsonschema_opt="disallow_bigints_as_strings" \
    --jsonschema_opt="prefix_schema_files_with_package" \
    --jsonschema_opt="json_fieldnames" \
    --doc_opt="json,proto-docs.json" \
    --doc2_opt="markdown,proto-docs.md" \
    --grpc-gateway_out v2/src \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt generate_unbound_methods=true \
    --openapiv2_out v2/src \
    --openapiv2_opt logtostderr=true \
    "${PROTOS[@]}" > /dev/stderr

pushd v2 > /dev/stderr
pushd src > /dev/stderr
gopatch ./... -p=/proto.patch
popd > /dev/stderr

gofmt -w . > /dev/stderr
find . -regextype egrep -regex ".*[.](go|json|md)$" -print0 | xargs -0 tar -cf -
