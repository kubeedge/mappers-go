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
# The root of the octopus directory
ROOT_DIR="${CURR_DIR}"
source "${ROOT_DIR}/hack/lib/init.sh"

function entry() {
  # copy template
  read -p "Please input the mapper name (like 'Bluetooth', 'BLE'): " -r mapperName
  if [[ -z "${mapperName}" ]]; then
    echo "the mapper name is required"
    exit 1
  fi
  mapperNameLowercase=$(echo -n "${mapperName}" | tr '[:upper:]' '[:lower:]')
  mapperPath="${ROOT_DIR}/mappers/${mapperNameLowercase}"
  if [[ -d "${mapperPath}" ]]; then
    echo "the directory is existed"
    exit 1
  fi
  cp -r "${ROOT_DIR}/_template/mapper" "${mapperPath}"

  mapperVar=$(echo "${mapperName}" | sed -e "s/\b\(.\)/\\u\1/g")
  sed -i "s/Template/${mapperVar}/g" `grep Template -rl ${mapperPath}`
  sed -i "s/mappers\/${mapperVar}/mappers\/${mapperNameLowercase}/g" `grep "mappers\/${mapperVar}" -rl ${ROOT_DIR}/mappers/${mapperNameLowercase}`
  sed -i "s/${mapperVar}/${mapperNameLowercase}/g" ${mapperPath}/Dockerfile
  # gofmt
  go fmt "${mapperPath}/..." >/dev/null 2>&1

  # build
  make -se mapper "${mapperNameLowercase}" build
}

entry "$@"
