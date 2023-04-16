CMDS = ssl-auto-recorder ssl-log-converter ssl-log-cutter ssl-log-indexer ssl-log-player ssl-log-recorder ssl-log-stats ssl-vision-tracker-client
DOCKER_TARGETS = $(addprefix docker-, $(CMDS))
.PHONY: all docker test install proto $(DOCKER_TARGETS)

all: install docker

docker: $(DOCKER_TARGETS)

$(DOCKER_TARGETS): docker-%:
	docker build --build-arg cmd=$* -t $*:latest .

test:
	go test ./...

install:
	go install -v ./...

proto:
	tools/generateProto.sh
