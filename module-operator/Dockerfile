# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

ARG BASE_IMAGE=ghcr.io/oracle/oraclelinux:8-slim
FROM $BASE_IMAGE AS build_base

# Need to use specific WORKDIR to match source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator
COPY . .

# Copy the operator binary
COPY bin/linux_amd64/verrazzano-module-operator /usr/local/bin/verrazzano-module-operator

RUN chmod 500 /usr/local/bin/verrazzano-module-operator

# Create the verrazzano-module-operator image
FROM scratch AS final

# Copy the operator binary
COPY --from=build_base --chown=1000 /usr/local/bin/verrazzano-module-operator /usr/local/bin/verrazzano-module-operator

WORKDIR /home/verrazzano
COPY --from=build_base --chown=1000 /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/manifests ./manifests
#COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/THIRD_PARTY_LICENSES.txt /licenses/

USER 1000

ENTRYPOINT ["/usr/local/bin/verrazzano-module-operator"]
