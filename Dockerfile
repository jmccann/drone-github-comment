# Docker image for the Drone GH PR Comment plugin
#
#     docker build -t jmccann/drone-github-comment .

#
# Run testing and build binary
#

FROM golang:1.10-alpine AS builder

# set working directory
RUN mkdir -p /go/src/github.com/jmccann/drone-github-comment
WORKDIR /go/src/github.com/jmccann/drone-github-comment

# copy sources
COPY . .

# run tests
RUN go test -v ./...

# build binary
RUN go build -v -o "/drone-github-comment"

#
# Build the image
#

FROM alpine:3.7

RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

COPY --from=builder /drone-github-comment /bin/drone-github-comment
ENTRYPOINT ["/bin/drone-github-comment"]
