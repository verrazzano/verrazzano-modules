# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

include make/global-env.mk

.PHONY: install-verrazzano-modules-operator
install-verrazzano-modules-operator:
	@echo "Running modules-operator install"
	${CI_SCRIPTS_DIR}/install_verrazzano_modules_operator.sh
