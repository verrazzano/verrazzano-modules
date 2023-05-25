# Project name

This repository contains the following content:

- [Verrazzano module operator](./module-operator) - a [Kubernetes operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) that can
  be deployed to a Kubernetes cluster, and install and uninstall Verrazzano modules from the cluster in which the operator is deployed.

## Installation

TBD

## Documentation

For instructions on using Verrazzano, see the [Verrazzano documentation](https://verrazzano.io/latest/docs/).

For detailed installation instructions, see the [Verrazzano Installation Guide](https://verrazzano.io/latest/docs/setup/install/installation/).

## Examples

TBD

## Help

See the [Verrazzano documentation](https://verrazzano.io/latest/) for how to join the Verrazzano
public Slack channel.

## Contributing

This project welcomes contributions from the community. Before submitting a pull request, please [review our contribution guide](./CONTRIBUTING.md)

## Testing

In order to run these make targets locally, copy tests/scripts/env-template.sh someplace and customize it to your liking.
The template file describes what minimum variables are needed to be able to run the targets.

To build verrazzano-module-operator and the deployment manifest

$ DOCKER_REPO=<Target Docker repo e.g ghcr.io> \
 DOCKER_NAMESPACE=<Target namespace for image e.g. verrazzano> \
 DOCKER_IMAGE_NAME=<Name for resulting image or leave blank for default> \
 DOCKER_CMD=<podman or leave empty for docker command> \
 DOCKER_CREDS_USR=<username for docker repo or empty> \
 DOCKER_CREDS_PSW=<password for docker repo or empty> \
 IMAGE_PULL_SECRETS=verrazzano-container-registry \
 make docker-push generate-operator-artifacts

To create a single KIND cluster:

$ make setup

To install verrazzano-modules-operator run

$ DOCKER_REPO=<Docker repo for the e.g ghcr.io> \
 DOCKER_CMD=<podman or leave empty for docker command> \
 DOCKER_CREDS_USR=<username for docker repo or empty> \
 DOCKER_CREDS_PSW=<password for docker repo or empty> \
 make install

To run a test, you can run any test suite using the TEST_SUITES variable and the "test" target:

$ TEST_SUITES="module-operator/verify-install/..." make test

## Security

Please consult the [security guide](./SECURITY.md) for our responsible security vulnerability disclosure process

## License

Copyright (c) 2023, Oracle and/or its affiliates.
