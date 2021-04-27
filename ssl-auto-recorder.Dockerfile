FROM golang:1.16-alpine AS build
WORKDIR /go/src/github.com/RoboCup-SSL/ssl-go-tools
COPY go.mod go.mod
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go install ./cmd/ssl-auto-recorder

# Start fresh from a smaller image
FROM alpine:3.9
COPY --from=build /go/bin/ssl-auto-recorder /app/ssl-auto-recorder
WORKDIR /data
ENTRYPOINT ["/app/ssl-auto-recorder"]
CMD []
