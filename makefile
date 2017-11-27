DOCKER_REPO_NAME:= gcr.io/npav-172917/
DOCKER_IMAGE_NAME := cisco-pipeline
GO_REPOSITORY_PATH := github.com/accedian/bigmuddy-network-telemetry-pipeline
DOCKER_VER := $(if $(DOCKER_VER),$(DOCKER_VER),dev)
BINARY=pipeline
BIN_PATH := bin/${BINARY}
GO_SDK_IMAGE := gcr.io/npav-172917/docker-go-sdk
GO_SDK_VERSION := 0.5.0-ubuntu16
 
GOPATH := $(GOPATH)

include skeleton/pipeline.mk

# Setup pretest as a prerequisite of tests.
test: pretest
pretest:
	@echo Setting up zookeeper, kafka. Docker required.
	tools/test/run.sh
dockerbin: .FORCE
	echo "PATH is $(GOPATH)"
	docker run -it --rm \
		-e GOPATH=/root/go \
		-v "$(GOPATH):/root/go" \
		-w "/root/go/src/$(GO_REPOSITORY_PATH)" \
		$(GO_SDK_IMAGE):$(GO_SDK_VERSION) make $(BIN_PATH) . 

docker: dockerbin
	 docker build -t $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER) .

push: docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)

circleci-binaries:
	go build -o $(BIN_PATH) .

circleci-push: circleci-docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)

circleci-docker: circleci-binaries
	docker build -t $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER) .
	
.FORCE: 
clean:  
	rm -rf bin
# name of executable.
    
