FROM golang:1.25-alpine@sha256:d3f0cf7723f3429e3f9ed846243970b20a2de7bae6a5b66fc5914e228d831bbb AS build_go
ARG cmd
WORKDIR /work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app"]
CMD []
