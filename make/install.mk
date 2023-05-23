# Copyright (C) 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

include ./global-env.mk
include ./env.mk

install-verrazzano-modules-operator: export INSTALL_CONFIG_FILE_KIND ?= ${TEST_SCRIPTS_DIR}/v1beta1/install-verrazzano-modules-kind.yaml
install-verrazzano-modules-operator: export POST_INSTALL_DUMP ?= false
.PHONY: install-verrazzano-modules-operator
install-verrazzano-modules-operator:
	@echo "Running KIND install"
	${CI_SCRIPTS_DIR}/install_verrazzano_modules_operator.sh
