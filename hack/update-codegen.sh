#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
GOPRIVATE="bitbucket.org/endclothing"

rm -rf "${SCRIPT_ROOT}"/pkg/client/

go mod vendor
chmod +x "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/generate-groups.sh

"${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/generate-groups.sh all \
github.com/endclothing/secret-manager-operator/pkg/client/ github.com/endclothing/secret-manager-operator/pkg/apis/ \
"endclothing.com:v1" \
--output-base "${SCRIPT_ROOT}" \
--go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

mv "${SCRIPT_ROOT}"/github.com/endclothing/secret-manager-operator/pkg/apis/endclothing.com/v1/zz_generated.deepcopy.go "${SCRIPT_ROOT}"/pkg/apis/endclothing.com/v1/
mv "${SCRIPT_ROOT}"/github.com/endclothing/secret-manager-operator/pkg/client "${SCRIPT_ROOT}"/pkg/client/

rm -rf "${SCRIPT_ROOT}"/github.com
rm -rf "${SCRIPT_ROOT}"/vendor
