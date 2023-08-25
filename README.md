# xvm

Xvm is a wrapper for popular sdk (Go/Node/...) written in Go. It automatically picks a good version of sdk given your current working directory, downloads it from the official server (if required) and then transparently passes through all command-line arguments to the real sdk binary.  You can call it just like you would call go/node/java/...

With xvm you can manage versions of the sdk in your code and keep scenarios like development and CI pipeline using the same version at all times, which helps with [reproducible build](https://reproducible-builds.org/).



In addition to this, you can also easily switch between different versions without the tedious installation process, which is very useful when you need to manage multiple projects working under different versions at the same time.

## Usage

### Installation

You can download `Xvm` binary on our [Releases](https://github.com/modern-devops/xvm/releases) page and add it to your PATH manually.

It is important to ensure that `Xvm` can be call using the command `xvm`.

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

### SDK Version

#### Go

rule: {env.XVM_GO_VERSION} > {project}/.goversion > {home}/.goversion > {project}/go.mod#version

If the environment variable `XVM_GO_VERSION` is set, it will use the version specified in the value.

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

rule: {env.XVM_NODE_VERSION} > {project}/.nodeversion > {home}/.nodeversion

If the environment variable `XVM_NODE_VERSION` is set, it will use the version specified in the value.

If the `.nodeversion` file in the root of your project has the following content, the version of `Node` will always be `20.5.1` in your project:

```text
20.5.1
```

If `.nodeversion` is not found in the project root directory, it will try to find it in the user's home.


## SDK Mirror

By default, Xvm retrieves sdk releases from official server.

For `Go`, the default mirror is https://go.dev/dl, overridden with the environment variable `XVM_GO_MIRROR`.

For `Node`, the default mirror is https://nodejs.org/dist, overridden with the environment variable `XVM_NODE_MIRROR`.

