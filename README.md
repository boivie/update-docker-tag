# Update Docker Tag

A simple utility to update the newly built docker image tag reference
in your kubernetes deployments.

Usage:

First set the environment variable `UPDATE_DOCKER_TAG_PATH` to where your
kubernetes templates are located. If this is not set, it will use
the current directory.

```
export UPDATE_DOCKER_TAG_PATH=~/path/to/k8s/
```

Then run `update-docker-tag <tag>`, providing the image tag that you just
built. If this is not specified, the latest docker tag will be found and
used.

The command will output which files it has updated. The discovery of images
to replace is quite good, but the patching is currently a bit basic, so
it could potentially do mistakes.
