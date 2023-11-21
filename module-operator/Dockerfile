# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
ARG BASE_IMAGE=ghcr.io/oracle/oraclelinux:8-slim

FROM ghcr.io/verrazzano/golang:v1.20.10 AS build_base

# Need to use specific WORKDIR to match source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-modules
COPY . .

# Build the operator binary
RUN go build -o /usr/local/bin/verrazzano-module-operator ./module-operator/main.go
RUN chmod 500 /usr/local/bin/verrazzano-module-operator

# Create the verrazzano-module-operator image
FROM $BASE_IMAGE AS final

# Copy the operator binary
COPY --from=build_base /usr/local/bin/verrazzano-module-operator /usr/local/bin/verrazzano-module-operator
COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/manifests/config/scripts/run.sh /
COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/manifests /manifests
COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/THIRD_PARTY_LICENSES.txt /licenses/

USER nobody

ENTRYPOINT ["/run.sh"]
