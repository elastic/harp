# Introduction

`harp` implements its own secret storage engine based on
an encrypted immutable monolithic file format.

Hopefully you don't have to manipulate a secret container at low level.
`harp` CLI tool allow you to handle secrets without the knowledge of the
container specification.

The container specification will be used by `harp` plugins for new feature
extension via external plugins. It defines the protocol and the format used to
transfer from point A to B, secrets and metadata associated.

---

* [Previous topic](../1-template-engine/9-usecases.md)
* [Index](../)
* [Next topic](2-specifications.md)
