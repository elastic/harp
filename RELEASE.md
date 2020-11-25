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
# Prepare docker image with all tools
mage docker:tools
# Release an artifact
RELEASE=v0.1.0 mage releaser:harp
```
