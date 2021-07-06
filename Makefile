.PHONY=all ssl-auto-recorder-docker ssl-log-player-docker

all: ssl-auto-recorder-docker ssl-log-player-docker

ssl-auto-recorder-docker:
	docker build -f ssl-auto-recorder.Dockerfile -t ssl-auto-recorder:latest .

ssl-log-player-docker:
	docker build -f ssl-log-player.Dockerfile -t ssl-log-player:latest .
