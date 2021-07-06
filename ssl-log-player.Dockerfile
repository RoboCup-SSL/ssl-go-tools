FROM golang:1.16-alpine AS build
WORKDIR /go/src/github.com/RoboCup-SSL/ssl-go-tools
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY internal internal
RUN go install ./cmd/ssl-log-player

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build /go/bin/ssl-log-player /app/ssl-log-player
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app/ssl-log-player"]
CMD []
