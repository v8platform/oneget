FROM golang:alpine3.13 AS builder

WORKDIR /src
COPY . .
RUN go build -o /out/oneget .

FROM alpine AS app
COPY --from=builder /out/oneget /usr/local/bin/oneget
ENTRYPOINT ["oneget"]