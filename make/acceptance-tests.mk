# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
export GO_TEST_ARGS ?= -v

run-test: export RANDOMIZE_TESTS ?= true
run-test: export RUN_PARALLEL ?= true
run-test: export TEST_REPORT ?= "test-report.xml"
run-test: export TEST_REPORT_DIR ?= "${WORKSPACE}/tests"
.PHONY: run-test
run-test:
	${CI_SCRIPTS_DIR}/run-go-tests.sh

run-test: export RANDOMIZE_TESTS := false
run-test: export RUN_PARALLEL := false
.PHONY: run-sequential
run-sequential: run-test

.PHONY: verify-operator
verify-operator:
	TEST_SUITES=verify-operator/... make test


test-reports: export TEST_REPORT ?= "test-report.xml"
test-reports: export TEST_REPORT_DIR ?= "${WORKSPACE}/tests"
.PHONY: test-reports
test-reports:
	# Copy the generated test reports to WORKSPACE to archive them
	mkdir -p ${TEST_REPORT_DIR}
	cd ${GO_REPO_PATH}/verrazzano-modules/tests
	find . -name "${TEST_REPORT}" | cpio -pdm ${TEST_REPORT_DIR}

.PHONY: pipeline-artifacts
pipeline-artifacts: test-reports

