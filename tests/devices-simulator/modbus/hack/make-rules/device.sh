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
ROOT_DIR="$(cd "${CURR_DIR}/.." && pwd -P)"

mkdir -p "${CURR_DIR}/bin"

function mod() {
  local device="${1}"

  # the device is sharing the vendor with root
  pushd "${CURR_DIR}" >/dev/null || exist 1
  echo "downloading dependencies for device ${device}..."

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
  [[ "${2:-}" != "only" ]] && mod "$@"
  local device="${1}"

  echo "fmt and linting device ${device}..."

  gofmt -s -w "${CURR_DIR}/"
  echo "${CURR_DIR}"
  golangci-lint run "${CURR_DIR}/..."

  echo "...done"
}

function build() {
  [[ "${2:-}" != "only" ]] && lint "$@"
  local device="${1}"

  local flags=" -w -s "
  local ext_flags=" -extldflags '-static' "
  local os="${OS:-$(go env GOOS)}"
  local arch="${ARCH:-$(go env GOARCH)}"

  local platform
  if [[ "${ARM:-false}" == "true" ]]; then
    echo "crossed packaging for linux/arm"
    platform=("linux/arm")
  elif [[ "${ARM64:-false}" == "true" ]]; then
    echo "crossed packaging for linux/arm64"
    platform=("linux/arm64")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platform=("${os}/${arch}")
  fi

  echo "building ${platform}"

  local os_arch
  IFS="/" read -r -a os_arch <<<"${platform}"
  local os=${os_arch[0]}
  local arch=${os_arch[1]}
  GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build \
    -ldflags "${flags} ${ext_flags}" \
    -o "${CURR_DIR}/bin/${device}_${os}_${arch}" \
    "${CURR_DIR}/cmd/main.go"

  cp ${CURR_DIR}/bin/${device}_${os}_${arch} ${CURR_DIR}/bin/${device}"Device"
  echo "...done"
}

function package() {
  [[ "${2:-}" != "only" ]] && build "$@"
  local device="${1}"

  echo "packaging device ${device}..."

  local image_name="${device}-simulator"
  local tag=v1.0

  local platform
  if [[ "${ARM:-false}" == "true" ]]; then
    echo "crossed packaging for linux/arm"
    platform=("linux/arm")
  elif [[ "${ARM64:-false}" == "true" ]]; then
    echo "crossed packaging for linux/arm64"
    platform=("linux/arm64")
  else
    local os="${OS:-$(go env GOOS)}"
    local arch="${ARCH:-$(go env GOARCH)}"
    platform=("${os}/${arch}")
  fi

  pushd "${CURR_DIR}" >/dev/null 2>&1
  if [[ "${platform}" =~ darwin/* ]]; then
    echo "package into Darwin OS image is unavailable, please use CROSS=true env to containerize multiple arch images or use OS=linux ARCH=amd64 env to containerize linux/amd64 image"
  fi

  local image_tag="${image_name}:${tag}-${platform////-}"
  echo "packaging ${image_tag}"
  sudo docker build \
    --platform "${platform}" \
    -t "${image_tag}" .
  popd >/dev/null 2>&1

  echo "...done"
}

function entry() {
  local device="${1:-}"
  shift 1

  local stages="${1:-build}"
  shift $(($# > 0 ? 1 : 0))

  IFS="," read -r -a stages <<<"${stages}"
  local commands=$*
  echo ${#stages[@]}
  if [[ ${#stages[@]} -ne 1 ]]; then
    commands="only"
  fi

  for stage in "${stages[@]}"; do
    echo "# make device ${device} ${stage} ${commands}"
    case ${stage} in
    m | mod) mod "${device}" "${commands}" ;;
    l | lint) lint "${device}" "${commands}" ;;
    b | build) build "${device}" "${commands}" ;;
    p | pkg | package) package "${device}" "${commands}" ;;
    *) echo "unknown action '${stage}', select from mod,lint,build package" ;;
    esac
  done

}

entry "$@"
