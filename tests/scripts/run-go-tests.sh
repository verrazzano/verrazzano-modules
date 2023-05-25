#!/usr/bin/env bash
# Copyright (c) 2022, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#
if [ ! -z "${kubeConfig}" ]; then
  export KUBECONFIG="${kubeConfig}"
fi
if [ -z "${TEST_SUITES}" ]; then
  echo "${0}: No test suites specified"
  exit 0
fi

TEST_ROOT=${TEST_ROOT:-"${GOPATH}/src/github.com/verrazzano/verrazzano-modules"}
TEST_DUMP_ROOT=${TEST_DUMP_ROOT:-"."}
SEQUENTIAL_SUITES=${SEQUENTIAL_SUITES:-false}

GO_TEST_ARGS=${GO_TEST_ARGS:-"-v"}
if [ "${RUN_PARALLEL}" == "true" ]; then
  GO_TEST_ARGS="${GO_TEST_ARGS} -p=1"
fi
if [ "${RANDOMIZE_TESTS}" == "true" ]; then
  GO_TEST_ARGS="${GO_TEST_ARGS} --shuffle=on"
fi
if [ -n "${TAGGED_TESTS}" ]; then
  GO_TEST_ARGS="${GO_TEST_ARGS} -tags=${TAGGED_TESTS}"
fi
####
#if [ -n "${INCLUDED_TESTS}" ]; then
#  GO_TEST_ARGS="${GO_TEST_ARGS} --focus-file=${INCLUDED_TESTS}"
#fi
#if [ -n "${EXCLUDED_TESTS}" ]; then
#  GO_TEST_ARGS="${GO_TEST_ARGS} --skip-file=${EXCLUDED_TESTS}"
#fi
#if [ -n "${DRY_RUN}" ]; then
#  GO_TEST_ARGS="${GO_TEST_ARGS} --dry-run"
#fi
if [ -n "${SKIP_DEPLOY}" ]; then
  TEST_ARGS="${TEST_ARGS} --skip-deploy=${SKIP_DEPLOY}"
fi
if [ -n "${SKIP_UNDEPLOY}" ]; then
  TEST_ARGS="${TEST_ARGS} --skip-undeploy=${SKIP_UNDEPLOY}"
fi

if [ -n "${TEST_ARGS}" ]; then
  TEST_ARGS="-- ${TEST_ARGS}"
fi

set -x
go test ${GO_TEST_ARGS} ${TEST_ROOT}/${TEST_SUITES} ${TEST_ARGS}