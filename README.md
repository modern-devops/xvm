# XVM

Xvm is a wrapper for popular sdk (Go/Node/...) written in Go. It automatically picks a good version of sdk given your current working directory, downloads it from the official server (if required) and then transparently passes through all command-line arguments to the real sdk binary.  You can call it just like you would call go/node/java/...

With xvm you can manage versions of the sdk in your code and keep scenarios like development and CI pipeline using the same version at all times, which helps with [reproducible build](https://reproducible-builds.org/).



In addition to this, you can also easily switch between different versions without the tedious installation process, which is very useful when you need to manage multiple projects working under different versions at the same time.

## Usage

### Installation

You can download `Xvm` binary on our [Releases](https://github.com/modern-devops/xvm/releases) page and add it to your PATH manually.

See the following to ensure that the command `xvm` is available.

For Mac/Linux:

```shell
# 1. Download the file to /path/to/xvm
curl -o /path/to/xvm -# -k {url}

# 2. Add an executable permission
$ chmod +x /path/to/xvm

# 3. Create a executable soft link
$ ln -s /path/to/xvm /usr/local/bin/xvm

# 4. Try to get help for xvm
$ xvm --help
```

For Windows:

```shell
# 1. Create a executable soft link
$ mklink C:\WINDOWS\system32\xvm.exe \to\path\xvm.exe

# 2. Try to get help for xvm
$ xvm --help
```

### Activate SDK

Activate Golang/Node/Java:

```shell
$ xvm activate go node java --add_binpath
```

You've got Golang/Node/Java that can be any version.

Open a new terminal and call the following command to experience the xvm-wrapped sdk:

```shell
# Uses the latest version of the go
$ go version

# Uses the latest version of the java
$ java -version

# Uses the latest version of the node
$ node --version

# Uses the latest version of the npm
$ npm --version
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

#### Java

rule: {env.XVM_JAVA_VERSION} > {project}/.javaversion > {home}/.javaversion

If the environment variable `XVM_JAVA_VERSION` is set, it will use the version specified in the value.

If the `.javaversion` file in the root of your project has the following content, the version of `Java` will always be `20.0.2` in your project:

```text
20.0.2
```

## SDK Mirror

By default, `Xvm` gets all available versions through the official indexing api and retrieves releases from official mirror.

Unless otherwise specified, with all SDKS you can override the indexing api with environment variable `XVM_{SDK}_API` and the mirror with environment variable `XVM_{SDK}_MIRROR`.

### Go

The default mirror is https://go.dev/dl, overridden with the environment variable `XVM_GO_MIRROR`.

If you do not configure the environment variable `XVM_GO_API`, the indexing api defaults to `{mirror}/?mode=json&include=all`

### Node

The default mirror is https://nodejs.org/dist, overridden with the environment variable `XVM_NODE_MIRROR`.

If you do not configure the environment variable `XVM_NODE_API`, the indexing api defaults to `{mirror}/index.json`

### Java

The default distribution is [zulu](https://www.azul.com/downloads/zulu-community/?package=jdk), overridden with the environment variable `XVM_JAVA_DISTRIBUTION`.

`Xvm` plans to support the following distribution of JDKS:

<!-- BEGIN GENERATED RESULTS TABLE -->

| Distribution | Description                      | Official site                                                               | License                                                                           | Supported? |
|--------------|----------------------------------|-----------------------------------------------------------------------------|-----------------------------------------------------------------------------------|------------|
| `zulu`       | Azul Zulu OpenJDK                | [Link](https://www.azul.com/downloads/zulu-community/?package=jdk)          | [Link](https://www.azul.com/products/zulu-and-zulu-enterprise/zulu-terms-of-use/) | ✔️         |
| `oracle`     | Oracle JDK                       | [Link](https://www.oracle.com/java/technologies/downloads/)                 | [Link](https://java.com/freeuselicense)                                           | ❌         |
| `temurin`    | Eclipse Temurin                  | [Link](https://adoptium.net/)                                               | [Link](https://adoptium.net/about.html)                                           | ❌         |
| `ms`         | Microsoft Build of OpenJDK       | [Link](https://www.microsoft.com/openjdk)                                   | [Link](https://docs.microsoft.com/java/openjdk/faq)                               | ❌         |
| `amazon`     | Amazon Corretto Build of OpenJDK | [Link](https://aws.amazon.com/corretto/)                                    | [Link](https://aws.amazon.com/corretto/faqs/)                                     | ❌         |
| `semeru`     | IBM Semeru Runtime Open Edition  | [Link](https://developer.ibm.com/languages/java/semeru-runtimes/downloads/) | [Link](https://openjdk.java.net/legal/gplv2+ce.html)                              | ❌         |


<!-- END GENERATED RESULTS TABLE -->

Regardless of the distribution, you can override the mirror with environment variable `XVM_JAVA_MIRROR` and the indexing API with environment variable `XVM_JAVA_INDEXING_API`

For `zulu`, the default mirror is https://cdn.azul.com/zulu/bin, the default indexing api is https://api.azul.com/zulu/download/community/v1.0/bundles/

## Show Detail

Run `xvm show [sdk]` to get more information about xvm.
