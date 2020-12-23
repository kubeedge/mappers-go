DESTDIR?=
USR_DIR?=/usr/local
INSTALL_DIR?=${DESTDIR}${USR_DIR}
INSTALL_BIN_DIR?=${INSTALL_DIR}/bin
GOPATH?=$(shell go env GOPATH)

# make all builds both cloud and edge binaries

BINARIES=cloudcore \
	admission \
	edgecore \
	edgesite \
	keadm


.EXPORT_ALL_VARIABLES:
OUT_DIR ?= _output

define ALL_HELP_INFO
# Build code.
#
# Args:
#   WHAT: binary names to build. support: $(BINARIES)
#         the build will produce executable files under $(OUT_DIR)
#         If not specified, "everything" will be built.
#
# Example:
#   make
#   make all
#   make all HELP=y
#   make all WHAT=cloudcore
#   make all WHAT=cloudcore GOLDFLAGS="" GOGCFLAGS="-N -l"
#     Note: Specify GOLDFLAGS as an empty string for building unstripped binaries, specify GOGCFLAGS
#     to "-N -l" to disable optimizations and inlining, this will be helpful when you want to
#     use the debugging tools like delve. When GOLDFLAGS is unspecified, it defaults to "-s -w" which strips
#     debug information, see https://golang.org/cmd/link for other flags.

endef
.PHONY: all
ifeq ($(HELP),y)
all: clean
	@echo "$$ALL_HELP_INFO"
else
all: verify-golang
endif


define VERIFY_HELP_INFO
# verify golang,vendor and codegen
#
# Example:
# make verify
endef
.PHONY: verify
ifeq ($(HELP),y)
verify:
	@echo "$$VERIFY_HELP_INFO"
else
verify:verify-golang verify-vendor verify-vendor-licenses
endif

.PHONY: verify-golang
verify-golang:
	hack/verify-golang.sh

.PHONY: verify-vendor
verify-vendor:
	hack/verify-vendor.sh
.PHONY: verify-vendor-licenses
verify-vendor-licenses:
	hack/verify-vendor-licenses.sh

define TEST_HELP_INFO
# run golang test case.
#
# Args:
#   WHAT: Component names to be testd. support: $(COMPONENTS)
#         If not specified, "everything" will be tested.
#
# Example:
#   make test
#   make test HELP=y
#   make test WHAT=cloud
endef
.PHONY: test
ifeq ($(HELP),y)
test:
	@echo "$$TEST_HELP_INFO"
else
test: clean
	hack/make-rules/test.sh $(WHAT)
endif

define LINT_HELP_INFO
# run golang lint check.
#
# Example:
#   make lint
#   make lint HELP=y
endef
.PHONY: lint
ifeq ($(HELP),y)
lint:
	@echo "$$LINT_HELP_INFO"
else
lint:
	hack/make-rules/lint.sh
endif



define E2E_HELP_INFO
# e2e test.
#
# Example:
#   make e2e
#   make e2e HELP=y
#
endef
.PHONY: e2e
ifeq ($(HELP),y)
e2e:
	@echo "$$E2E_HELP_INFO"
else
e2e:
#	bash tests/e2e/scripts/execute.sh device_crd
#	This has been commented temporarily since there is an issue of CI using same master for all PRs, which is causing failures when run parallelly
	tests/e2e/scripts/execute.sh
endif

define KEADM_E2E_HELP_INFO
# keadm e2e test.
#
# Example:
#   make keadm_e2e
#   make keadm_e2e HELP=y
#
endef
.PHONY: keadm_e2e
ifeq ($(HELP),y)
keadm_e2e:
	@echo "KEADM_E2E_HELP_INFO"
else
keadm_e2e:
	tests/e2e/scripts/keadm_e2e.sh
endif

define CLEAN_HELP_INFO
# Clean up the output of make.
#
# Example:
#   make clean
#   make clean HELP=y
#
endef
.PHONY: clean
ifeq ($(HELP),y)
clean:
	@echo "$$CLEAN_HELP_INFO"
else
clean:
	hack/make-rules/clean.sh
endif


QEMU_ARCH ?= x86_64
ARCH ?= amd64
IMAGE_TAG ?= $(shell git describe --tags  --always)
GO_LDFLAGS='$(shell hack/make-rules/version.sh)'


# Mappers
.PHONY: bluetoothmapper
bluetoothmapper: clean
	hack/make-rules/bluetoothmapper.sh
.PHONY: bluetoothmapper_image
bluetoothmapper_image:bluetoothmapper
	sudo docker build -t bluetoothmapper:v1.0 ./mappers/bluetooth-go/

.PHONY: modbusmapper
modbusmapper: clean
	hack/make-rules/modbusmapper.sh
.PHONY: modbusmapper_image
modbusmapper_image:modbusmapper
	sudo docker build -t modbusmapper:v1.0 ./mappers/modbus-go

.PHONY: mappers
mappers:bluetoothmapper modbusmapper

