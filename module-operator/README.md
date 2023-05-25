[![Go Report Card](https://goreportcard.com/badge/github.com/verrazzano/verrazzano-module-operator)](https://goreportcard.com/report/github.com/verrazzano/verrazzano-module-operator)

# Verrazzano Module Operator

Instructions for building and testing the Verrazzano module operator.

## Prerequisites
* `controller-gen` v0.9.2
* `go` version v1.19
* Docker
* `kubectl`
* `helm` (for testing)

## Build the Verrazzano module operator

* To build the operator:

    ```
    $ make go-build
    ```

## Build and push Docker image

* To build the Docker image:
    ```
    $ DOCKER_REPO=<repo> DOCKER_NAMESPACE=<namespace> make docker-build
    ```

* To build and push the Docker image:
    ```
    $ DOCKER_REPO=<repo> DOCKER_NAMESPACE=<namespace> make docker-push
    ```

* To build and push the Docker image and generate artifacts:
    ```
    $ IMAGE_PULL_SECRETS=<secret name> DOCKER_REPO=<repo> DOCKER_NAMESPACE=<namespace> make generate-operator-artifacts docker-push
    ```

## Running on a cluster

1. After building and pushing the Docker image and generating the operator artifacts, apply the operator YAML:

    ```sh
    $ kubectl apply -f build/deploy/verrazzano-module-operator.yaml
    ```

2. Wait for the operator:

    ```sh
    $ kubectl -n verrazzano-install rollout status deployment/verrazzano-module-operator
    ```

3. Alternatively, apply the CRDs and use `make` to run the operator:

    ```sh
    $ kubectl apply -f manifests/charts/operators/verrazzano-module-operator/crds/*
    make run
    ```

## Testing

1. After installing the operator, apply a Module CR:

    ```sh
    $ kubectl apply -f - <<EOF
    apiVersion: platform.verrazzano.io/v1alpha1
    kind: Module
    metadata:
      name: vz-test
    spec:
      moduleName: helm
      targetNamespace: default
    EOF
    ```

2. Verify the Helm chart was installed:

    ```sh
    $ helm ls
    ```

## Modifying the API definitions
If you update the API definitions, you must regenerate CRDs and code.

* To generate manifests (for example, CRDs):

    ```
    $ make manifests
    ```

* To generate code (for example, `zz_generated.deepcopy.go`):

    ```
    $ make generate
    ```
