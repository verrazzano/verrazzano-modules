# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

FROM ghcr.io/verrazzano/verrazzano-base:v1.0.0-20230327155846-4653b27 AS build_base

# Need to use specific WORKDIR to match source packages
WORKDIR /root/go/src/github.com/verrazzano/verrazzano-modules/helm-operator
COPY . .

# Copy the operator binary
COPY bin/linux_amd64/verrazzano-helm-operator /usr/local/bin/verrazzano-helm-operator

RUN chmod 500 /usr/local/bin/verrazzano-helm-operator

# Create the verrazzano-helm-operator image
FROM ghcr.io/verrazzano/verrazzano-base:v1.0.0-20230327155846-4653b27 AS final

RUN groupadd -r verrazzano \
    && useradd --no-log-init -r -m -d /verrazzano -g verrazzano -u 1000 verrazzano \
    && mkdir /home/verrazzano \
    && chown -R 1000:verrazzano /home/verrazzano

COPY --from=build_base --chown=verrazzano:verrazzano /usr/local/bin/verrazzano-helm-operator /usr/local/bin/verrazzano-helm-operator

# Copy the operator binary
WORKDIR /home/verrazzano
COPY --from=build_base --chown=verrazzano:verrazzano /root/go/src/github.com/verrazzano/verrazzano-modules/helm-operator/manifests/config/scripts/run.sh .
#COPY --from=build_base /root/go/src/github.com/verrazzano/verrazzano-modules/THIRD_PARTY_LICENSES.txt /licenses/

USER 1000

ENTRYPOINT ["/home/verrazzano/run.sh"]