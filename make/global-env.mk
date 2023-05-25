# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

export GOPATH ?= ${HOME}/go
export GO_REPO_PATH ?= ${GOPATH}/src/github.com/verrazzano

export VM_ROOT := ${GO_REPO_PATH}/verrazzano-modules
export VMO_ROOT := ${VM_ROOT}/module-operator
export CI_ROOT ?= ${VM_ROOT}
export CI_SCRIPTS_DIR ?= ${CI_ROOT}/tests/scripts

export WORKSPACE ?= ${HOME}/verrazzano-modules-workspace
export KUBECONFIG ?= ${WORKSPACE}/test_kubeconfig

export IMAGE_PULL_SECRET ?= verrazzano-container-registry
export DOCKER_REPO ?= 'ghcr.io'
export DOCKER_NAMESPACE ?= 'verrazzano'
export TEST_ROOT ?= ${VM_ROOT}/tests
export TEST_SCRIPTS_DIR ?= ${TEST_ROOT}/scripts
export OPERATOR_YAML ?= ${VMO_ROOT}/build/deploy/verrazzano-module-operator.yaml