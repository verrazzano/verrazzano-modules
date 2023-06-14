# Copyright (C) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

export DOCKER_CMD ?= docker
##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

#.PHONY: vet
#vet: ## Run go vet against code.
#	go vet ./...

##@ Build

#
# Go build related tasks
#
.PHONY: go-build
go-build:
	$(GO) build \
		-ldflags "${GO_LDFLAGS}" \
		-o bin/$(shell uname)_$(shell uname -m)/${NAME} \
		main.go

.PHONY: go-build-linux
go-build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build \
		-ldflags "-s -w ${GO_LDFLAGS}" \
		-o bin/linux_amd64/${NAME} \
		main.go

.PHONY: go-build-linux-debug
go-build-linux-debug:
	GOOS=linux GOARCH=amd64 $(GO) build \
		-ldflags "${GO_LDFLAGS}" \
		-o out/linux_amd64/${NAME} \
		main.go

.PHONY: go-install
go-install: fmt
	$(GO) install ./...

.PHONY: build
build: fmt go-build

#                                                                                                                                                         # Test-related tasks                                                                                                                                      #                                                                                                                                                         .PHONY: unit-test
unit-test: go-install
	$(GO) test -v ${TEST_PATHS}

.PHONY: install-crds
install-crds:
	kubectl apply -f charts/operators/verrazzano-module-operator/crd

.PHONY: run
run: ## Run a controller from your host.
	$(GO) run ./main.go

# If you wish built the manager image targeting other platforms you can use the --platform flag.
# (i.e. ${DOCKER_CMD} build --platform linux/arm64 ). However, you must enable ${DOCKER_CMD} buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: docker-build-common

docker-build-common: BASE_IMAGE ?= ghcr.io/verrazzano/verrazzano-base:v1.0.0-20230529051534-037adf4-4653b27@sha256:73c0b081e13228fbbe5fb87b46cde1781cba25bdf9cf6489daa56460e5e1435b
.PHONY: docker-build-common
docker-build-common:
	@echo Building ${NAME} image ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG}
	# the TPL file needs to be copied into this dir so it is in the ${DOCKER_CMD} build context
	#cp ../THIRD_PARTY_LICENSES.txt .
	${DOCKER_CMD} buildx build --pull --platform linux/arm64,linux/amd64 -f Dockerfile \
		--build-arg BASE_IMAGE=${BASE_IMAGE} \
		-t ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ..

.PHONY: docker-push
docker-push: docker-build docker-push-common

.PHONY: docker-push-debug
docker-push-debug: docker-build-debug docker-push-common

.PHONY: docker-push-common
docker-push-common:
	${DOCKER_CMD} tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG}
ifdef DOCKER_CREDS_USR
ifdef DOCKER_CREDS_PSW
	@${DOCKER_CMD} login ${DOCKER_REPO} --username ${DOCKER_CREDS_USR} --password ${DOCKER_CREDS_PSW}
endif
endif
	$(call retry_docker_push,${DOCKER_IMAGE_FULLNAME}:${DOCKER_IMAGE_TAG})
ifeq ($(CREATE_LATEST_TAG), "1")
	${DOCKER_CMD} tag ${DOCKER_IMAGE_NAME}:${DOCKER_IMAGE_TAG} ${DOCKER_IMAGE_FULLNAME}:latest;
	$(call retry_docker_push,${DOCKER_IMAGE_FULLNAME}:latest);
endif

.PHONY: generate-operator-artifacts
generate-operator-artifacts:
	mkdir -p ${BUILD_DEPLOY} ; \
	env DOCKER_IMAGE_FULLNAME=${DOCKER_IMAGE_FULLNAME} DOCKER_IMAGE_TAG=${DOCKER_IMAGE_TAG} \
		CHART_NAME=${NAME} \
		MODULE_ROOT=${WORKING_DIR} \
		IMAGE_PULL_SECRETS=${IMAGE_PULL_SECRETS} \
		BUILD_OUT=${BUILD_DEPLOY} \
		OPERATOR_YAML=${OPERATOR_YAML} \
		../tools/scripts/generate_operator_artifacts.sh

