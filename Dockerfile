FROM golang:1.20-alpine AS build_go
ARG cmd
WORKDIR work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app/${cmd}
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENV COMMAND="/app/${cmd}"
ENTRYPOINT "${COMMAND}"
CMD []
