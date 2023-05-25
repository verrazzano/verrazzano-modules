# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

include make/global-env.mk

export CLUSTER_NAME ?= kind

setup-kind: export KUBERNETES_CLUSTER_VERSION ?= 1.24
.PHONY: setup-kind
setup-kind:
	@echo "Setup KIND cluster"
	${CI_SCRIPTS_DIR}/setup_kind.sh

#clean-kind: export KUBECONFIG ?= "${WORKSPACE}/test_kubeconfig"
.PHONY: delete-kind
delete-kind:
	@echo "Deleting kind cluster ${CLUSTER_NAME}, KUBECONFIG=${KUBECONFIG}"
	kind delete cluster --name=${CLUSTER_NAME}
	rm -rf ${KUBECONFIG}

.PHONY: delete-kind-all
delete-kind-all:
	@echo "Deleting all kind clusters"
	kind delete clusters --all
