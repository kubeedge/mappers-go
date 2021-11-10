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

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

function clean() {
  containerID=$(docker ps -a -f ancestor=opcua-simulator:v1.0-linux-amd64 -q)
  if [[ "$containerID" != "" ]]
  then
    docker stop "$containerID" && docker rm "$containerID"
  fi
  containerID2=$(docker ps -a -f ancestor=opcua-mapper:v1.0-linux-amd64 -q)
  if [[ "$containerID2" != "" ]]
  then
    docker stop "$containerID2" && docker rm "$containerID2"
  fi
}

function mod() {
  local e2e="opcua"

  # the device is sharing the vendor with root
  pushd "${CURR_DIR}" >/dev/null || exist 1
  echo "downloading dependencies for e2e ${e2e}..."

  if [[ "$(go env GO111MODULE)" == "off" ]]; then
    echo "go mod has been disabled by GO111MODULE=off"
  else
    echo "tidying"
    go mod tidy
    echo "vending"
    go mod vendor
  fi

  echo "...done"
  popd >/dev/null || return
}

function lint() {
  [[ "${2:-}" != "only" ]] && mod
  local e2e="opcua"

  echo "fmt and linting e2e ${e2e}..."

  gofmt -s -w "${CURR_DIR}"
  cd "${CURR_DIR}"
  golangci-lint run "${CURR_DIR}/..."

  echo "...done"
}

function start_test() {
  [[ "${2:-}" != "only" ]] && lint
  local e2e="opcua"

  echo "run e2e test ${e2e}..."

  go test -v

  echo "...done"
}

function entry() {
  local e2e="${1:-}"
  shift 1

  local stages="${1:-test}"
  shift $(($# > 0 ? 1 : 0))

  IFS="," read -r -a stages <<<"${stages}"
  local commands=$*
  if [[ ${#stages[@]} -ne 1 ]]; then
    commands="only"
  fi

  for stage in "${stages[@]}"; do
    echo "# make e2e ${e2e} ${stage} ${commands}"
    case ${stage} in
    m | mod) mod "${e2e}" "${commands}" ;;
    l | lint) lint "${e2e}" "${commands}" ;;
    t | test) start_test "${e2e}" "${commands}" ;;
    *) echo "unknown action '${stage}', select from mod,lint,test" ;;
    esac
  done

}


set -Ee
clean
entry "$@"

