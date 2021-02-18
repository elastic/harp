# How to release new version

## Prepare git tag

```sh
git tag -asm "<release message>" cmd/<tool>/v<version>
git push --tags
```

Example

```sh
git tag -asm "Cloud features" cmd/harp/v0.1.0
git push --tags
```

## Prepare artifacts

```sh
# Prepare docker image with all tools and go toolchain
mage docker:tools
# Create the docker image with all artifacts
RELEASE=v0.1.0 mage releaser:harp
# Extract artifact from image
docker run -ti --rm --volume $(pwd)/bin:/dist elastic/harp:artifacts-v0.1.0
$ cp * ../dist
# Pack release binaries
RELEASE=v0.1.0 task
```
