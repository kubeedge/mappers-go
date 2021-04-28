#!/usr/bin/env bash

###
#Copyright 2021 The KubeEdge Authors.
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.
###

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
ROOT_DIR="${CURR_DIR}"
source "${ROOT_DIR}/hack/lib/init.sh"


function find_subdirs() {
  local path="$1"
  if [[ -z "$path" ]]; then
    path="./"
  fi
  ls -l "$path" -I "common" | grep "^d" | awk '{print $NF}'
}

function entry() {
  local action="${1:-build}"

  shift $(($# > 0 ? 1 : 0))

  for d in $(find_subdirs "${CURR_DIR}/mappers"); do
    "${ROOT_DIR}/hack/make-rules/mapper.sh" "${d}" "${action}" "$@"
  done
}

echo $@
entry "$@"
