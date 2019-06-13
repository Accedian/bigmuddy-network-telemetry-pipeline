DOCKER_REPO_NAME:= gcr.io/npav-172917/
DOCKER_IMAGE_NAME := cisco-pipeline
GO_REPOSITORY_PATH := github.com/cisco/bigmuddy-network-telemetry-pipeline
DOCKER_VER := $(if $(DOCKER_VER),$(DOCKER_VER),$*(shell whoami)-dev)
DOCKER_SDK:= gcr.io/npav-172917/docker-go-sdk:0.12.0-ubuntu16
BINARY=pipeline
BIN_PATH := bin/${BINARY}
DOCKERFILE=docker/Dockerfile

include skeleton/pipeline.mk

dockerbin:
	docker run -it --rm \
		-v "$(GOPATH):/root/go" \
		-w "/root/go/src/$(GO_REPOSITORY_PATH)/" \
		$(DOCKER_SDK) make $(BIN_PATH) . 

docker: dockerbin
	mkdir -p docker/bin
	cp bin/${BINARY} docker/bin/
	docker build -f $(DOCKERFILE) -t $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER) .

push: docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)

circleci-push: circleci-docker
	docker push $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER)

circleci-docker: circleci-binaries
	mkdir -p docker/bin
	cp bin/${BINARY} docker/bin/
	docker build -f $(DOCKERFILE) -t $(DOCKER_REPO_NAME)$(DOCKER_IMAGE_NAME):$(DOCKER_VER) .

circleci-binaries: .FORCE
	make $(BIN_PATH) .

.FORCE: 
clean:  
	rm -rf bin
	rm -rf docker/bin

    
