# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

ARG BASE_IMAGE=ghcr.io/oracle/oraclelinux:8-slim
FROM $BASE_IMAGE AS build_base

# Need to use specific WORKDIR to match source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-modules/helm-operator
COPY . .

# Copy the operator binary
COPY bin/linux_amd64/verrazzano-helm-operator /usr/local/bin/verrazzano-helm-operator

RUN chmod 500 /usr/local/bin/verrazzano-helm-operator

# Create the verrazzano-helm-operator image
FROM $BASE_IMAGE AS final

RUN groupadd -r verrazzano \
    && useradd --no-log-init -r -m -d /verrazzano -g verrazzano -u 1000 verrazzano \
    && mkdir /home/verrazzano \
    && chown -R 1000:verrazzano /home/verrazzano

# Set up the Operator Node on Oracle Linux 8
RUN microdnf install oracle-olcne-release-el8 && \
    dnf config-manager --enable ol8_olcne16 ol8_addons ol8_baseos_latest ol8_appstream ol8_kvm_appstream ol8_UEKR7 && \
    dnf config-manager --disable ol8_olcne15 ol8_olcne14 ol8_olcne13 ol8_olcne12 ol8_developer || true && \
    microdnf install -y olcne-prometheus-chart

COPY --from=build_base --chown=verrazzano:verrazzano /usr/local/bin/verrazzano-helm-operator /usr/local/bin/verrazzano-helm-operator

# Copy the operator binary
WORKDIR /home/verrazzano
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano-modules/helm-operator/manifests/config/scripts/run.sh .
#COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/THIRD_PARTY_LICENSES.txt /licenses/

USER 1000

ENTRYPOINT ["/home/verrazzano/run.sh"]