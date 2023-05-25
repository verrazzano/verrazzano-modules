#!/usr/bin/env bash
# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
# Required env vars:
# WORKSPACE - workspace for output files, temp files, etc
# TEST_SCRIPTS_DIR - Location of the E2E tests directory
# KUBECONFIG - kubeconfig path for the target cluster
# GO_REPO_PATH - Local path to the Verrazzano Github repo
#
set -o pipefail

if [ -z "$WORKSPACE" ]; then
  echo "WORKSPACE must be set"
  exit 1
fi
if [ -z "$TEST_SCRIPTS_DIR" ]; then
  echo "TEST_SCRIPTS_DIR must be set to the E2E test script directory location"
  exit 1
fi
if [ -z "${KUBECONFIG}" ]; then
  echo "KUBECONFIG must be set"
  exit 1
fi
if [ -z "${VMO_ROOT}" ]; then
  echo "VMO_ROOT must be set"
  exit 1
fi

set -e
if [ -n "${IMAGE_PULL_SECRET}" ] && [ -n "${DOCKER_REPO}" ] && [ -n "${DOCKER_CREDS_USR}" ] && [ -n "${DOCKER_CREDS_PSW}" ]; then
  echo "Create Image Pull Secrets"
  # Create the verrazzano-install namespace
  kubectl create namespace verrazzano-install || true
  # create secret in verrazzano-install ns
  ${TEST_SCRIPTS_DIR}/create-image-pull-secret.sh "${IMAGE_PULL_SECRET}" "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}" "verrazzano-install"
fi

TARGET_OPERATOR_FILE=${TARGET_OPERATOR_FILE:-"${WORKSPACE}/verrazzano-modules-operator.yaml"}
if [ -z "$VZ_MODULES_OPERATOR_YAML" ]; then
  # copy the file, then install else ask to generate
  if [ -f "${VMO_ROOT}/build/deploy/verrazzano-module-operator.yaml" ]; then
    echo "Using pre-generated verrazzano-modules-operator.yaml"
    cp "${VMO_ROOT}/build/deploy/verrazzano-module-operator.yaml" ${TARGET_OPERATOR_FILE}
  else
    echo "verrazzano-module-operator.yaml does not exists, please generate one using make  generate-operator-artifacts"
    exit 1
  fi
else
  # The verrazzano-modules-operator.yaml filename was provided, install using that file.
  echo "Using provided operator.yaml file: " ${VZ_MODULES_OPERATOR_YAML}
  TARGET_OPERATOR_FILE=${VZ_MODULES_OPERATOR_YAML}
fi

echo "Installing Verrazzano Modules Operator on Kind"
if [ -f "${TARGET_OPERATOR_FILE}" ]; then
  kubectl apply -f ${TARGET_OPERATOR_FILE}
else
  echo "${TARGET_OPERATOR_FILE} does not exist"
fi

result=$?
if [[ $result -ne 0 ]]; then
  exit 1
fi

exit 0
