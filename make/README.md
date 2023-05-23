In order to run these make targets locally, copy tests/scripts/env-template.sh someplace and customize it to your liking.
The template file describes what minimum variables are needed to be able to run the targets.

To build verrazzano-module-operator and the deployment manifest

$ DOCKER_REPO=<Target Docker repo e.g ghcr.io> \
 DOCKER_NAMESPACE=<Target namespace for image e.g. verrazzano> \
 DOCKER_IMAGE_NAME=<Name for resulting image or leave blank for default> \
 DOCKER_CMD=<podman or leave empty for docker command> \
 DOCKER_CREDS_USR=<username for docker repo or empty> \
 DOCKER_CREDS_PSW=<password for docker repo or empty> \
 make precommit check-repo-clean docker-push generate-operator-artifacts

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
