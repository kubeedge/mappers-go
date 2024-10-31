#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

CURR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
ROOT_DIR="$(cd "${CURR_DIR}/../.." && pwd -P)"
source "${ROOT_DIR}/hack/lib/init.sh"

mkdir -p "${CURR_DIR}/bin"
mkdir -p "${CURR_DIR}/dist"

function mod() {
  [[ "${2:-}" != "only" ]]
  local mapper="${1}"

  # the mapper is sharing the vendor with root
  pushd "${ROOT_DIR}" >/dev/null || exist 1
  echo "downloading dependencies for mapper ${mapper}..."

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
  local mapper="${1}"

  echo "fmt and linting mapper ${mapper}..."

  gofmt -s -w "${CURR_DIR}/"
  golangci-lint run "${CURR_DIR}/..."

  echo "...done"
}

function build() {
  [[ "${2:-}" != "only" ]] && lint "$@"
  local mapper="${1}"

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

  if [[ "${arch}" == "amd64" ]]; then
    sudo sed -i '$a export GENICAM_GENTL64_PATH='${CURR_DIR}'/baumer/Ubuntu-16.04/x86_64' /etc/profile
    sed -i '$a export GENICAM_GENTL64_PATH='${CURR_DIR}'/baumer/Ubuntu-16.04/x86_64' ~/.bashrc
    sudo sed -i '1a '${CURR_DIR}'/bin/genicam/Linux64_x64' /etc/ld.so.conf
    cp ${CURR_DIR}/bin/librcapi_linux_x64.so ${CURR_DIR}/bin/librcapi.so
    cp ${CURR_DIR}/Dockerfile.x64 ${CURR_DIR}/Dockerfile
  elif [[ "${arch}" == "aarch64" ]]; then
    sudo sed -i '$a export GENICAM_GENTL64_PATH='${CURR_DIR}'/baumer/Ubuntu-16.04/arm64' /etc/profile
    sed -i '$a export GENICAM_GENTL64_PATH='${CURR_DIR}'/baumer/Ubuntu-16.04/arm64' ~/.bashrc
    sudo sed -i '1a ${CURR_DIR}/bin/genicam/Linux64_ARM' /etc/ld.so.conf
    cp ${CURR_DIR}/bin/librcapi_aarch64.so ${CURR_DIR}/bin/librcapi.so
    cp ${CURR_DIR}/Dockerfile.arm64 ${CURR_DIR}/Dockerfile
  else
    echo "can't find arch command or your arch is unsupport!"
    return
  fi
  sudo sysctl -w net.core.rmem_max=33554432
  sudo sysctl -w net.core.netdev_max_backlog=2000
  sudo sysctl -w net.core.netdev_budget=600

  local os_arch
  IFS="/" read -r -a os_arch <<<"${platform}"
  local os=${os_arch[0]}
  local arch=${os_arch[1]}
  GOOS=${os} GOARCH=${arch} CGO_ENABLED=1 go build \
    -ldflags "${flags} ${ext_flags}" \
    -o "${CURR_DIR}/bin/${mapper}_${os}_${arch}" \
    "${CURR_DIR}/cmd/main.go"

  cp ${CURR_DIR}/bin/${mapper}_${os}_${arch} ${CURR_DIR}/bin/${mapper}
  echo "...done"
}

function package() {
  [[ "${2:-}" != "only" ]] && build "$@"
  local mapper="${1}"

  echo "packaging mapper ${mapper}..."

  local image_name="${mapper}-mapper"
  if [[ -n "$2" ]]; then
    local tag=$2
  else
    local tag=v1.0
  fi

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

function test() {
  [[ "${2:-}" != "only" ]] && build "$@"
  local mapper="${1}"

  echo "running unit tests for mapper ${mapper}..."

  local unit_test_targets=(
    "${CURR_DIR}/config/..."
    "${CURR_DIR}/configmap/..."
    "${CURR_DIR}/device/..."
    "${CURR_DIR}/driver/..."
  )

  local os="${OS:-$(go env GOOS)}"
  local arch="${ARCH:-$(go env GOARCH)}"
  if [[ "${arch}" == "arm" ]]; then
    # NB(thxCode): race detector doesn't support `arm` arch, ref to:
    # - https://golang.org/doc/articles/race_detector.html#Supported_Systems
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=1 go test \
      -cover -coverprofile "${CURR_DIR}/dist/coverage_${mapper}_${os}_${arch}.out" \
      "${unit_test_targets[@]}"
  else
    GOOS=${os} GOARCH=${arch} CGO_ENABLED=1 go test \
      -race \
      -cover -coverprofile "${CURR_DIR}/dist/coverage_${mapper}_${os}_${arch}.out" \
      "${unit_test_targets[@]}"
  fi

  echo "...done"
}

function clean() {
  local mapper="${1}"

  echo "cleanup mapper ${mapper}..."

  rm -rf ${CURR_DIR}/bin/gige*  
  sudo sed -i '/GENICAM_GENTL/'d /etc/profile
  sudo sed -i '/GENICAM_GENTL/'d ~/.bashrc
  sudo sed -i '/bin\/genicam\/Linux/'d /etc/ld.so.conf
  sudo ldconfig
  rm -f ${CURR_DIR}/bin/librcapi.so ${CURR_DIR}/Dockerfile
  
  echo "...done"
}

function entry() {
  local mapper="${1:-}"
  shift 1

  local stages="${1:-build}"
  shift $(($# > 0 ? 1 : 0))

  IFS="," read -r -a stages <<<"${stages}"
  local commands=$*
  if [[ ${#stages[@]} -ne 1 ]]; then
    commands="only"
  fi

  for stage in "${stages[@]}"; do
    echo "# make mapper ${mapper} ${stage} ${commands}"
    case ${stage} in
    m | mod) mod "${mapper}" "${commands}" ;;
    l | lint) lint "${mapper}" "${commands}" ;;
    b | build) build "${mapper}" "${commands}" ;;
    p | pkg | package) package "${mapper}" "${commands}" ;;
    t | test) test "${mapper}" "${commands}" ;;
    c | clean) clean "${mapper}" "${commands}" ;;
    *) echo "unknown action '${stage}', select from mod,lint,build,test,clean" ;;
    esac
  done
}

echo $@
entry "$@"
