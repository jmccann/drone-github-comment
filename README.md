# drone-github-comment

Drone plugin to add comments to a Github Issues/PRs.

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo
docker build --rm=true -t jmccann/drone-github-comment .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-github-comment' not found or does not exist.
```

## Usage

Execute from the working directory:

```
docker run --rm \
  jmccann/drone-github-comment:1 --repo-owner jmccann --repo-name drone-github-comment \
  --pull-request 12 --api-key abcd1234 --message "Hello World!"
```
