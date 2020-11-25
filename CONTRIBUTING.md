# How to contribute

This document outlines some of the conventions on development workflow, commit
message formatting, contact points and other resources to make it easier to get
your contribution accepted.

## Getting started

- Fork the repository on GitHub.
- Read the README.md for build instructions.
- Play with the project, submit bugs, submit patches!

## Contribution flow

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where you want to base your work. This is usually master.
- Make commits of logical units and add test case if the change fixes a bug or
  adds new functionality.
- Run tests and make sure all the tests are passed.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- Submit a pull request to the repo.

Thanks for your contributions!

## Coding Style

The coding style suggested by the Golang community is used in this repository. See the
[style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

## Commit message guidelines

### Structure of the commit message

```text
<type>(optional scope): <description>

optional body

optional footer
```

### **Type**

- **fix**: The commit represents a bug fix for the bridge components
- **feat**: The commit represents a new feature for the bridge components
- **doc**: The commit represents a change in the documentation
- **chore**: Other changes

The type **fix** and **feat** will be added to changelog and version will be bumped.

> Please note that using `BREAKING CHANGE:` will bump the `major` version.

### **Optional scope**

- `fix` and `feat` use the name of command. Example: `fix(vctx): <description>`
- `doc`: try to use name of command. Example: `doc(harp): <description>`
- `chore`: readme, build scripts, dependencies, ..

### **Description**

The description contains succinct description of the change:

- use the imperative, present tense: "change" not "changed" nor "changes"
- don't capitalize first letter
- no dot (.) at the end

### **Optional body**

Just as in the subject, use the imperative, present tense: "change" not "changed"
nor "changes". The body should include the motivation for the change and contrast
this with previous behavior. This can also reference issues.

### **Optional footer**

All breaking changes have to be mentioned in the footer. You can also reference issues.

## Examples

```text
feat(harp): add virtual filesystem

implementd using Afero library

close #149
```

```text
chore: add contributing guidelines
```
