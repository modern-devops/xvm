# xvm

Xvm is a wrapper for popular sdk (Go/Node/...) written in Go. It automatically picks a good version of sdk given your current working directory, downloads it from the official server (if required) and then transparently passes through all command-line arguments to the real sdk binary.  You can call it just like you would call go/node/java/...

## Usage

### Installation

### Activate SDK

Activate Golang/Node:

```shell
$ xvm activate go node --add_binpath
```

You've got Golang and Node that can be any version.

Open a new terminal and call the following command to experience the xvm-wrapped sdk:

```shell
# Uses the latest version of the go
$ go version

# Uses the latest version of the node
$ node version

# Uses the latest version of the npm
$ npm version
```

### About Version

#### Go

rule: {project}/.goversion > {home}/.goversion > {project}/go.mod#version

If the `.goversion` file in the root of your project has the following content, the version of `Go` will always be 1.21.0 in your project:

```text
1.21.0
```

If `.goversion` is not found in the project root directory, it will try to find it in the user's home.

If the `go.mod` file in the root of your project has the following content, the version of `Go` will always be 1.21.0 in your project:

```text
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

