# Docker image for the Drone GH PR Comment plugin
#
#     docker build -t jmccann/drone-gh-pr-comment .

FROM alpine:3.5

RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

ADD drone-gh-pr-comment /bin/
ENTRYPOINT ["/bin/drone-gh-pr-comment"]
