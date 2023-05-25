#!/usr/bin/env bash
#
# Copyright (c) 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#

set -o pipefail

#set -xv

if [ -z "$GO_REPO_PATH" ]; then
  echo "GO_REPO_PATH must be set"
  exit 1
fi
if [ -z "$WORKSPACE" ]; then
  echo "WORKSPACE must be set"
  exit 1
fi
if [ -z "$TEST_SCRIPTS_DIR" ]; then
  echo "TEST_SCRIPTS_DIR must be set to the E2E test script directory location"
  exit 1
fi

scriptHome=$(dirname ${BASH_SOURCE[0]})

set -e

export KUBECONFIG=${KUBECONFIG:-"${WORKSPACE}/test_kubeconfig"}
export KUBERNETES_CLUSTER_VERSION=${KUBERNETES_CLUSTER_VERSION:-"1.22"}

CONNECT_JENKINS_RUNNER_TO_NETWORK=false
if [ -n "${JENKINS_URL}" ]; then
  echo "Running in Jenkins, URL=${JENKINS_URL}"
  CONNECT_JENKINS_RUNNER_TO_NETWORK=true
fi

CLUSTER_NAME=${CLUSTER_NAME:="kind"}
KIND_NODE_COUNT=${KIND_NODE_COUNT:-1}
KIND_CACHING=${KIND_CACHING:="false"}
KIND_NODE_COUNT=${KIND_NODE_COUNT:-1}

mkdir -p $WORKSPACE || true

echo "Create Kind cluster"
${scriptHome}/create_kind_clusters.sh "${CLUSTER_NAME}" "${KUBECONFIG}" "${KUBERNETES_CLUSTER_VERSION}" true ${CONNECT_JENKINS_RUNNER_TO_NETWORK} ${KIND_AT_CACHE} "NONE" ${KIND_NODE_COUNT}
if [ $? -ne 0 ]; then
  mkdir -p $WORKSPACE/kind-logs$()
  kind export logs $WORKSPACE/kind-logs
  echo "Kind cluster creation failed"
  exit 1
fi

kubectl wait --for=condition=ready nodes/${CLUSTER_NAME}-control-plane --timeout=5m --all
kubectl wait --for=condition=ready pods/kube-controller-manager-${CLUSTER_NAME}-control-plane -n kube-system --timeout=5m
echo "Listing pods in kube-system namespace ..."
kubectl get pods -n kube-system

exit 0
