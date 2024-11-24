FROM golang:1.22.5 AS builder
WORKDIR /work

COPY ./go.mod ./go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
RUN go build -o ccregisternotifier

FROM ubuntu:latest
RUN apt-get update && apt-get install -y ca-certificates curl --no-install-recommends && rm -rf /var/lib/apt/lists/*

COPY --from=builder /work/ccregisternotifier /usr/local/bin

CMD ["ccregisternotifier"]
