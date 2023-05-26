# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

.DEFAULT_GOAL := help
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

include make/quality.mk
include make/global-env.mk
include make/acceptance-tests.mk
include make/kind.mk
include make/install.mk

SCRIPT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))/tools/scripts
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

ifneq "$(MAKECMDGOALS)" ""
ifeq ($(MAKECMDGOALS),$(filter $(MAKECMDGOALS),docker-push))
ifndef DOCKER_REPO
    $(error DOCKER_REPO must be defined as the name of the docker repository where image will be pushed)
endif
ifndef DOCKER_NAMESPACE
    $(error DOCKER_NAMESPACE must be defined as the name of the docker namespace where image will be pushed)
endif
endif
endif

TIMESTAMP := $(shell date -u +%Y%m%d%H%M%S)
DOCKER_IMAGE_TAG ?= local-${TIMESTAMP}-$(shell git rev-parse --short HEAD)
VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME ?= module-operator-dev

VERRAZZANO_MODULE_OPERATOR_IMAGE = ${DOCKER_REPO}/${DOCKER_NAMESPACE}/${VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME}:${DOCKER_IMAGE_TAG}

CURRENT_YEAR = $(shell date +"%Y")

PARENT_BRANCH ?= origin/main

GO ?= CGO_ENABLED=0 GO111MODULE=on GOPRIVATE=github.com/verrazzano go
GO_LDFLAGS ?= -extldflags -static -X main.buildVersion=${BUILDVERSION} -X main.buildDate=${BUILDDATE}

.PHONY: clean
clean: ## remove coverage and test results
	find . -name coverage.cov -exec rm {} \;
	find . -name coverage.html -exec rm {} \;
	find . -name coverage.raw.cov -exec rm {} \;
	find . -name \*-test-result.xml -exec rm {} \;
	find . -name coverage.xml -exec rm {} \;
	find . -name unit-test-coverage-number.txt -exec rm {} \;

#@ Build

.PHONY: docker-build
docker-build: ## build and push all images
	(cd module-operator; make docker-build DOCKER_IMAGE_NAME=${VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG})

.PHONY: docker-push
docker-push: ## build and push all images
	(cd module-operator; make docker-push DOCKER_IMAGE_NAME=${VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG})

.PHONY: docker-push-debug
docker-push-debug: ## build and push all images
	(cd module-operator; make docker-push-debug DOCKER_IMAGE_NAME=${VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG})

.PHONY: generate-operator-artifacts
generate-operator-artifacts: ## build and push all images
	(cd module-operator; make generate-operator-artifacts DOCKER_IMAGE_NAME=${VERRAZZANO_MODULE_OPERATOR_IMAGE_NAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG})

#@ Testing

.PHONY: precommit
precommit: precommit-check precommit-build unit-test-coverage ## run all precommit checks

.PHONY: precommit-nocover
precommit-nocover: precommit-check precommit-build unit-test ## run precommit checks without code coverage check

.PHONY: precommit-check
precommit-check: check check-tests copyright-check ## run precommit checks without unit testing

.PHONY: precommit-build
precommit-build:  ## go build the project
	go build ./...

.PHONY: unit-test-coverage
unit-test-coverage:  ## run unit tests with coverage
	${SCRIPT_DIR}/coverage.sh html

.PHONY: unit-test-coverage-ratcheting
unit-test-coverage-ratcheting:  ## run unit tests with coverage ratcheting
	${SCRIPT_DIR}/coverage-number-comparison.sh

.PHONY: unit-test
unit-test:  ## run all unit tests in project
	go test $$(go list ./...)


#@  Compliance check targets


#@ Compliance

.PHONY: fix-copyright
fix-copyright: ## run fix-copyright from the verrazzano repo
	go run github.com/verrazzano/verrazzano-modules/tools/copyright .

.PHONY: copyright-check-year
copyright-check-year: ## check copyright notices have correct current year
	go run github.com/verrazzano/verrazzano-modules/tools/copyright . --enforce-current $(shell git log --since=01-01-${CURRENT_YEAR} --name-only --oneline --pretty="format:" | sort -u)

.PHONY: copyright-check
copyright-check: copyright-check-year  ## check copyright notices are correct
	go run github.com/verrazzano/verrazzano-modules/tools/copyright .

.PHONY: copyright-check-local
copyright-check-local:  ## check copyright notices are correct in local working copy
	go run github.com/verrazzano/verrazzano-modules/tools/copyright --verbose --enforce-current  $(shell git status --short | cut -c 4-)

.PHONY: copyright-check-branch
copyright-check-branch: copyright-check ## check copyright notices are correct in parent branch
	go run github.com/verrazzano/verrazzano-modules/tools/copyright --verbose --enforce-current $(shell git diff --name-only ${PARENT_BRANCH})

#
# Quality checks on acceptance tests
#

#@ Quality

.PHONY: check-tests
check-tests: check-eventually ## check test code for known quality issues

.PHONY: check-eventually
check-eventually: ## check for correct use of Gomega Eventually func
	#go run github.com/verrazzano/verrazzano-modules/tools/eventually-checker tests/e2e

.PHONY: check-repo-clean
check-repo-clean:
	(cd module-operator; make check-repo-clean)

# Create a kind cluster
.PHONY: setup
setup:
	@echo "Creating kind cluster."
	make setup-kind

# Peforms an install of module-operator to the target system
.PHONY: install
install:
	@echo "Installing verrazzano-modules-operator."
	make install-verrazzano-modules-operator

# Run tests
test: export TEST_SUITES ?= ./...
.PHONY: test
test:
	@echo "Running tests ${TEST_SUITES}"
	make run-test

# Remove the kind cluster
.PHONY: cleanup
cleanup:
	@echo "Removing kind cluster."
	make delete-kind

# Remove all the kind clusters
.PHONY: cleanup-all
cleanup-all:
	@echo "Removing all kind clusters."
	make delete-kind-all

