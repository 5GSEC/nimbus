# Want to contribute?

Great! We welcome contributions of all kinds, big or small! This includes bug reports, code fixes, documentation
improvements, and code examples.

Before you dive in, please take a moment to read through this guide.

# Reporting issue

We use [GitHub](https://github.com/5GSEC/nimbus) to manage the issues. Please open
a [new issue](https://github.com/5GSEC/nimbus/issues/new) directly there.

# Getting Started

## Setting Up Your Environment

- Head over to [GitHub](https://github.com/5GSEC/nimbus) and fork the 5GSec Nimbus repository.
- Clone your forked repository onto your local machine.
  ```shell
  git clone git@github.com:<your-username>/nimbus.git
  ```

## Install development tools

You'll need these tools for a smooth development experience:

- [Make](https://www.gnu.org/software/make/#download)
- [Go](https://go.dev/doc/install) SDK, version 1.21 or later
- Go IDE ([Goland](https://www.jetbrains.com/go/) / [VS Code](https://code.visualstudio.com/download))
- Container tools ([Docker](https://www.docker.com/) / [Podman](https://podman.io/))
- [Kubernetes cluster](https://kubernetes.io/docs/setup/) running version 1.26 or later.
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) version 1.26 or later.

# Project Setup

## Building Locally

- Install development tools (mentioned above).

- Build Nimbus using:
  ```shell
  make build
  ```

## Testing Local Build

### Against the Cluster (without installing as workload):

#### Nimbus operator

- Generate code and manifests:
  ```shell
  make manifests generate
  ```

- Install CRDs:
  ```shell
  make install
  ```

- Run the operator:
  ```shell
  make run
  ```

#### Adapters

- Navigate to adapter's directory:
  ```shell
  cd pkg/adapter/<adapter-name>
  ```
- Run it:
  ```shell
  make run
  ```

### In the Cluster (installing as workload):

Follow [this](deployments/nimbus/Readme.md) guide to install Nimbus or the complete suite.

Alternatively, follow [this](docs/adapters.md) guide to install individual adapters.

# Contributing Code

### Understanding the Project

Before contributing to any Open Source project, it's important to have basic understanding of what the project is about.
It is advised to try out the project as an end user.

### Pull Requests and Code Reviews

We use GitHub [pull requests](https://github.com/5GSEC/nimbus/pulls) for code contributions. All submissions, including
those from project members, require review before merging.
We typically aim for two approvals per pull request, with reviews happening within a week or two.
Feel free to ping reviewers if you haven't received feedback within that timeframe.

#### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for clear and
consistent commit messages.

Please make sure you have added the **Signed-off-by:** footer in your git commit. In order to do it automatically, use
the **--signoff** flag:

```shell
git commit --signoff
```

With this command, git would automatically add a footer by reading your name and email from git config.

# Testing and Documentation

Tests and documentation are not optional, make sure your pull requests include:

- Tests that verify your changes and don't break existing functionality.
- Updated [documentation](docs) reflecting your code changes.
- Reference information and any other relevant details.

## Commands to run tests

- Integration tests:
  ```shell
  make integration-test
  ```

- End-to-end tests:
  **Requires installing the complete suite, follow [this](deployments/nimbus/Readme.md)**
  ```shell
  make e2e-test
  ```
