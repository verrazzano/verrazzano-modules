#!/usr/bin/env bash
#
# Copyright (c) 2020, 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#

SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)

IMAGE_PULL_SECRETS=${IMAGE_PULL_SECRETS:-}
BUILD_OUT=${BUILD_OUT:-${SCRIPT_DIR}/../../build/deploy}

set -ueo pipefail

command -v helm >/dev/null 2>&1 || {
  fail "helm is required but cannot be found on the path. Aborting."
}

function check_helm_version {
    local helm_version=$(helm version --short | cut -d':' -f2 | tr -d " ")
    local major_version=$(echo $helm_version | cut -d'.' -f1)
    local minor_version=$(echo $helm_version | cut -d'.' -f2)
    if [ "$major_version" != "v3" ]; then
        echo "Helm version is $helm_version, expected v3!" >&2
        return 1
    fi
    return 0
}

check_helm_version || exit 1

IMAGE_PULL_SECRET_ARG=
if [ -n "${IMAGE_PULL_SECRETS}" ] ; then
    IMAGE_PULL_SECRET_ARG="--set imagePullSecrets={${IMAGE_PULL_SECRETS}}"
fi

set -x
CHARTS_OUT=${BUILD_OUT}/charts
rm -rf ${CHARTS_OUT}
mkdir -p ${BUILD_OUT}
cp -pr $SCRIPT_DIR/../../manifests/charts ${BUILD_OUT}
TARGET_CHART=${CHARTS_OUT}/verrazzano-module-operator
TARGET_VALUES=${TARGET_CHART}/values.yaml

if [ -n "${IMAGE_PULL_SECRETS}" ] ; then
  yq -i eval '.imagePullSecrets[0].name = env(IMAGE_PULL_SECRETS)' ${TARGET_VALUES}
fi
yq -i eval '.image.repository = env(DOCKER_IMAGE_FULLNAME)'  ${TARGET_VALUES}
yq -i eval '.image.tag = env(DOCKER_IMAGE_TAG)' ${TARGET_VALUES}
helm package ${TARGET_CHART} -d ${BUILD_OUT}

helm template --include-crds ${TARGET_CHART} -n "verrazzano-install" > ${OPERATOR_YAML}

exit $?
