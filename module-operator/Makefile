# Copyright (C) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GO ?= GO111MODULE=on GOPRIVATE=github.com/verrazzano go

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

NAME:=verrazzano-module-operator
REPO_NAME:=verrazzano-module-operator

CONTROLLER_GEN_VERSION ?= $(shell go list -m -f '{{.Version}}' sigs.k8s.io/controller-tools)
CREATE_LATEST_TAG=0

CRD_OPTIONS ?= "crd:crdVersions=v1,maxDescLen=0"
KUBECONFIG ?= ${HOME}/.kube/config

ifndef DOCKER_IMAGE_FULLNAME
DOCKER_IMAGE_NAME ?= ${NAME}-dev
DOCKER_IMAGE_FULLNAME=${DOCKER_IMAGE_NAME}
ifeq ($(MAKECMDGOALS),$(filter $(MAKECMDGOALS),docker-push push-tag))
	ifndef DOCKER_REPO
		$(error DOCKER_REPO must be defined as the name of the docker repository where image will be pushed)
	endif
	ifndef DOCKER_NAMESPACE
		$(error DOCKER_NAMESPACE must be defined as the name of the docker namespace where image will be pushed)
	endif
endif
ifdef DOCKER_NAMESPACE
DOCKER_IMAGE_FULLNAME := ${DOCKER_NAMESPACE}/${DOCKER_IMAGE_FULLNAME}
endif
ifdef DOCKER_REPO
DOCKER_IMAGE_FULLNAME := ${DOCKER_REPO}/${DOCKER_IMAGE_FULLNAME}
endif
endif

DOCKER_IMAGE_TAG ?= local-$(shell git rev-parse --short HEAD)

SHORT_COMMIT_HASH ?= $(shell git rev-parse --short=8 HEAD)

OPERATOR_IMAGE ?= ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	# Add copyright headers to the kubebuilder generated CRDs
	./hack/add-crd-header.sh
	./hack/update-codegen.sh "platform:v1alpha1" "boilerplate.go.txt"

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

#.PHONY: vet
#vet: ## Run go vet against code.
#	go vet ./...

##@ Build

#
# Go build related tasks
#
.PHONY: go-build
go-build:
	$(GO) build \
		-ldflags "${GO_LDFLAGS}" \
		-o bin/$(shell uname)_$(shell uname -m)/${NAME} \
		main.go

.PHONY: go-build-linux
go-build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build \
		-ldflags "-s -w ${GO_LDFLAGS}" \
		-o bin/linux_amd64/${NAME} \
		main.go

.PHONY: go-build-linux-debug
go-build-linux-debug:
	GOOS=linux GOARCH=amd64 $(GO) build \
		-ldflags "${GO_LDFLAGS}" \
		-o out/linux_amd64/${NAME} \
		main.go

.PHONY: go-install
go-install:
	$(GO) install ./...

.PHONY: build
build: fmt go-build

#                                                                                                                                                         # Test-related tasks                                                                                                                                      #                                                                                                                                                         .PHONY: unit-test
unit-test: go-install
	$(GO) test -v  ./... ./controllers/... ./apis/...

.PHONY: run
run: ## Run a controller from your host.
	$(GO) run ./main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64 ). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: go-build ## Build docker image with the manager.
	docker build -t ${OPERATOR_IMAGE} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${OPERATOR_IMAGE}

# find or download controller-gen
CONTROLLER_GEN_PATH := $(shell eval go env GOPATH)
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell command -v controller-gen))
	$(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION}
	$(eval CONTROLLER_GEN=$(CONTROLLER_GEN_PATH)/bin/controller-gen)
else
	$(eval CONTROLLER_GEN=$(shell command -v controller-gen))
endif
	@{ \
	set -eu; \
	ACTUAL_CONTROLLER_GEN_VERSION=$$(${CONTROLLER_GEN} --version | awk '{print $$2}') ; \
	if [ "$${ACTUAL_CONTROLLER_GEN_VERSION}" != "${CONTROLLER_GEN_VERSION}" ] ; then \
		echo  "Bad controller-gen version $${ACTUAL_CONTROLLER_GEN_VERSION}, please install ${CONTROLLER_GEN_VERSION}" ; \
		exit 1; \
	fi ; \
	}

