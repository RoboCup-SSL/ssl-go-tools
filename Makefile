CMDS = ssl-auto-recorder ssl-log-converter ssl-log-cutter ssl-log-indexer ssl-log-player ssl-log-recorder ssl-log-stats ssl-vision-tracker-client ssl-multicast-sources
DOCKER_TARGETS = $(addprefix docker-, $(CMDS))
.PHONY: all docker test install proto $(DOCKER_TARGETS)

all: install docker

docker: $(DOCKER_TARGETS)

$(DOCKER_TARGETS): docker-%:
	docker build --build-arg BINARY_NAME=$* -t $*:latest .

test:
	go test ./...

install:
	go install -v ./...

proto:
	buf generate

update-go:
	go get -v -u all

update: update-go proto
