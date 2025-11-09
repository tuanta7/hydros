# Integration Tests

This directory contains integration tests for the Hydros OAuth2/OIDC server using `dockertest` to spin up real
dependency containers.

```shell
go test -v ./test/...
```

## Prerequisites

- Docker must be installed and running

## Troubleshooting

Docker Desktop uses a different socket path. If tests fail to connect

```shell
# Check your Docker context
docker context ls

# If using Docker Desktop, ensure it's the active context
docker context use desktop-linux

# Or set DOCKER_HOST explicitly before running tests
export DOCKER_HOST=unix://$HOME/.docker/desktop/docker.sock
go test -v ./test/...
```