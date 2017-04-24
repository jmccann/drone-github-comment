# drone-gh-pr-comment

Drone plugin to add comments to a Github PR.

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
docker build --rm=true -t jmccann/drone-gh-pr-comment .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-gh-pr-comment' not found or does not exist.
```

## Usage

Execute from the working directory:

```
docker run --rm \
  jmccann/drone-gh-pr-comment:1 --repo-owner jmccann --repo-name drone-gh-pr-comment \
  --pull-request 12 --api-key abcd1234 --message "Hello World!"
```
