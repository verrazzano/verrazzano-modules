# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
ARG BASE_IMAGE=ghcr.io/oracle/oraclelinux:9-slim

FROM ghcr.io/verrazzano/oraclelinux-golang:v1.20.12 AS build_base

# Need to use specific WORKDIR to match source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-modules
COPY . .

RUN dnf install -y oraclelinux-developer-release-el8 && \
    dnf config-manager --enable ol8_developer && \
    dnf update -y && \
    dnf install -y go-toolset-1.19.6

# Build the operator binary
RUN go build -o /usr/local/bin/verrazzano-module-operator ./module-operator/main.go
RUN chmod 500 /usr/local/bin/verrazzano-module-operator

# Create the verrazzano-module-operator image
FROM $BASE_IMAGE AS final

RUN microdnf update -y && \
    microdnf clean all && \
    rm -rf /var/cache/yum/*

RUN groupadd -r verrazzano \
    && useradd --no-log-init -r -m -d /verrazzano -g verrazzano -u 1000 verrazzano \
    && mkdir /home/verrazzano \
    && chown -R 1000:verrazzano /home/verrazzano

# Copy the operator binary
COPY --from=build_base --chown=verrazzano:verrazzano /usr/local/bin/verrazzano-module-operator /usr/local/bin/verrazzano-module-operator

WORKDIR /home/verrazzano
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/manifests/config/scripts/run.sh .
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/manifests ./manifests
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano-modules/module-operator/THIRD_PARTY_LICENSES.txt /licenses/

USER 1000

ENTRYPOINT ["/home/verrazzano/run.sh"]
