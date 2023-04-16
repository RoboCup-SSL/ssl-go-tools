.PHONY: all docker docker-ssl-auto-recorder docker-ssl-log-player test install proto

all: install docker

docker: docker-ssl-auto-recorder docker-ssl-log-player

docker-ssl-auto-recorder:
	docker build -f ./cmd/ssl-auto-recorder/Dockerfile -t ssl-auto-recorder:latest .

docker-ssl-log-player:
	docker build -f ./cmd/ssl-log-player/Dockerfile -t ssl-log-player:latest .

test:
	go test ./...

install:
	go install -v ./...

proto:
	tools/generateProto.sh
