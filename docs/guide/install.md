---
order: 1
description: Steps to install Starport on your local computer.
---

[Install Starport](#install-starport)
  - [Prerequisites](#prerequisites)
    - [Operating Systems](#operating-systems)
    - [Go](#go)
  - [Verify Your Starport Version](#verify-your-starport-version)
  - [Installing Starport](#installing-starport)
    - [Write permission](#write-permission)
  - [Upgrading Your Starport Installation](#upgrading-your-starport-installation)
  - [Installing Starport on macOS with Homebrew](#installing-starport-on-macos-with-homebrew)
  - [Build from source](#build-from-source)
  - [Summary](#summary)


# Install Starport

You can run [Starport](https://github.com/tendermint/starport) in a web-based Gitpod IDE or you can install Starport on your local computer. 


## Prerequisites

Be sure you have met the prerequisites before you install and use Starport. 

### Operating Systems

Starport is supported for the following operating systems:

- GNU/Linux
- macOS
- Windows Subsystem for Linux (WSL)

### Go 

Starport is written in the Go programming language. To use Starport on a local system:

- Install [Go](https://golang.org/doc/install) (**version 1.16** or higher)
- Ensure the Go environment variables are [set properly](https://golang.org/doc/gopath_code#GOPATH) on your system

## Verify Your Starport Version 

To verify the version of Starport you have installed, run the following command:

```sh
starport version
```

## Installing Starport

To install the latest version of the `starport` binary use the following command.

```bash
curl https://get.starport.network/starport! | bash
```

This command invokes `curl` to download the install script and pipes the output to `bash` to perform the installation. The `starport` binary is installed in `/usr/local/bin`.

To learn more or customize the installation process, see [Starport installer docs](https://github.com/allinbits/starport-installer) on GitHub.

### Write Permission

Starport installation requires write permission to the `/usr/local/bin/` directory. If the installation fails because you do not have write permission to `/usr/local/bin/`, run the following command:

```bash
curl https://get.starport.network/starport | bash
```

Then run this command to move the `starport` executable to `/usr/local/bin/`:

```bash
sudo mv starport /usr/local/bin/
```

## Upgrading Your Starport Installation

Before you install a new version of Starport, remove all existing Starport installations. 

To remove the current Starport installation:

1. On your terminal window, press `Ctrl+C` to stop the chain that you started with `starport chain serve`.
1. Remove the Starport binary with `rm $(which starport)`.
   Depending on your user permissions, run the command with or without `sudo`.
1. Repeat this step until all `starport` installations are removed from your system.

After all existing Starport installations are removed, follow the [Installing Starport with cURL](#installing-starport-with-curl) instructions. For details on version features and changes, see the [changelog.md](https://github.com/tendermint/starport/blob/develop/changelog.md) in the repo.

## Installing Starport on macOS with Homebrew

Using brew to install Starport is supported only for macOS machines without the M1 chip. 

```bash
brew install tendermint/tap/starport
```

To install Starport on macOS machines with the M1 chip, use the `curl` command as described in [Installing Starport](#installing-starport). 

## Build from source

```bash
git clone https://github.com/tendermint/starport --depth=1
cd starport && make install
```

## Summary

- Verify the prerequisites.
- To setup a local development environment, install Starport locally on your computer.
- Install Starport by fetching the binary using cURL, Homebrew, or by building from source.
- The latest version is installed by default. You can install previous versions of the precompiled `starport` binary.
- Stop the chain and remove existing versions before installing a new version.
