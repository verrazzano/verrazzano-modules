# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

# Set WORKSPACE where you want all temporary test files to go, etc
# - default is ${HOME}/verrazzano-workspace
#export WORKSPACE=

# Set this to true to enable script debugging
#export VZ_TEST_DEBUG=true

# Used for the Github packages repo creds for the image pull secret for private branch builds
#export DOCKER_CREDS_USR=my-github-user
#export DOCKER_CREDS_PSW=$(cat ~/.github_token)
#export DOCKER_REPO=ghcr.io

# Override where the Kubeconfig for the cluster is stored
#export KUBECONFIG= # Default is ${WORKSPACE}/test_kubeconfig

# Location of the Verrazzano Module Operator manifest
#export VZ_MODULES_OPERATOR_YAML= #Default is ${VMO_ROOT}/build/deploy/verrazzano-module-operator.yaml
