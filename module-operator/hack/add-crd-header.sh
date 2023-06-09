#!/bin/bash
# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

# Add YAML boilerplate to generated CRDs - kubebuilder currently does not seem to have a way to
# add boilerplate headers to these - only to generated Go files

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
GENERATED_CRDS_DIR=$SCRIPT_DIR/../manifests/charts/operators/verrazzano-module-operator/crds

# The following two steps are required to handle the cases of running "make manifests" when there
# are and are not api changes.  This is necessary because fix-copyright currently cannot handle both
# cases correctly with the same set of options.

# First put in the headers from the Git history
go run github.com/verrazzano/verrazzano-modules/tools/fix-copyright -useExistingUpdateYearFromHeader $GENERATED_CRDS_DIR

# Then fix the updated year for files that were modified this year
go run github.com/verrazzano/verrazzano-modules/tools/fix-copyright $GENERATED_CRDS_DIR

