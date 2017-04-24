# Docker image for the Drone GH PR Comment plugin
#
#     docker build -t jmccann/drone-github-comment .

FROM alpine:3.5

RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

ADD drone-github-comment /bin/
ENTRYPOINT ["/bin/drone-github-comment"]
