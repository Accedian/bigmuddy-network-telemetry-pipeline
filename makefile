DOCKER_REPO_NAME:= gcr.io/npav-172917/
DOCKER_IMAGE_NAME := cisco-pipeline
GO_REPOSITORY_PATH := github.com/cisco/bigmuddy-network-telemetry-pipeline
DOCKER_VER := $(if $(DOCKER_VER),$(DOCKER_VER),dev)
BINARY=pipeline
BIN_PATH := bin/${BINARY}
GO_SDK_IMAGE := gcr.io/npav-172917/docker-go-sdk
GO_SDK_VERSION := 0.5.0-ubuntu16
DOCKERFILE=docker/Dockerfile

ifeq ($(CIRCLECI),true)
	GOPATH := /root/go
endif

include skeleton/pipeline.mk

bling:
	echo $(GOPATH)
# Setup pretest as a prerequisite of tests.
test: pretest
pretest:
	@echo Setting up zookeeper, kafka. Docker required.
	tools/test/run.sh
dockerbin: .FORCE
	docker run -it --rm \
	 	-v "$(GOPATH):/root/go" \
	 	-w "/root/go/src/$(GO_REPOSITORY_PATH)" \
	 	$(GO_SDK_IMAGE):$(GO_SDK_VERSION) make $(BIN_PATH) . 

docker: dockerbin
	mkdir -p docker/bin
	cp bin/${BINARY} docker/bin/
	docker build -f $(DOCKERFILE) -t $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER) .

push: docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)


circleci-push: docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)

.FORCE: 
clean:  
	rm -rf bin
	rm -rf docker/bin

    
