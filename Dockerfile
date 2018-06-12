# build stage
FROM golang:alpine AS build-env
# Install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh && \
    apk add build-base

# Test and build muapi
ADD . /go/src/github.com/woosteln/mqclientspawner
RUN go get -u github.com/golang/dep/cmd/dep
RUN cd /go/src/github.com/woosteln/mqclientspawner && \
  dep ensure && \
  CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/mqclientspawner

# final stage
FROM alpine:3.7
RUN apk update && apk upgrade && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=build-env /go/src/github.com/woosteln/mqclientspawner/bin/mqclientspawner /usr/bin/mqclientspawner

RUN chmod +x /usr/bin/mqclientspawner
ENTRYPOINT ["/usr/bin/mqclientspawner"]
CMD ["/usr/bin/mqclientspawner"]