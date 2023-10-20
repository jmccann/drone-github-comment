FROM alpine:3.7

RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

COPY drone-github-comment /bin/drone-github-comment
ENTRYPOINT ["/bin/drone-github-comment"]
