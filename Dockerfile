FROM alpine:latest
COPY oneget /
ENTRYPOINT ["/oneget"]