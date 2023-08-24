# xvm
xvm is a tool that provides version management for multiple sdks, such as golang/nodejs, allowing you to dynamically specify a version through a version file without installation.

## Usage

### Activate SDK

1. Activate Golang/Node SDK

```shell
$ xvm activate go node --add_binpath
```

2. Open a new terminal

```shell
# Uses the latest version of the go
$ go version

# Uses the latest version of the node
$ node version

# Uses the latest version of the npm
$ npm version
```

### Version AsCode

#### Go

rule: {project}/.goversion > {home}/.goversion > {project}/go.mod#version

If the `.goversion` file in the root of your project has the following content, the version of `Go` will always be 1.21.0 in your project:

```text
1.21.0
```

If `.goversion` is not found in the project root directory, it will try to find it in the user's home.

If the `go.mod` file in the root of your project has the following content, the version of `Go` will always be 1.21.0 in your project:

```text
module github.com/modern-devops/xvm

go 1.21.0
```

Note that before 1.21.0, the version in the `go.mod` file allowed only two sets of numbers, such as 1.17, so the `Go` version in the project would also be 1.17.

#### Node

rule: {project}/.nodeversion > {home}/.nodeversion

If the `.nodeversion` file in the root of your project has the following content, the version of `Node` will always be `20.5.1` in your project:

```text
20.5.1
```

If `.nodeversion` is not found in the project root directory, it will try to find it in the user's home.

### Show Detail

```shell
xvm show
```
