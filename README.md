# Argus

[![GitHub](https://img.shields.io/github/license/release-argus/argus)](https://github.com/release-argus/Argus/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/release-argus/Argus)](https://goreportcard.com/report/github.com/release-argus/Argus)
[![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/release-argus/argus?filename=go.mod)](https://go.dev/dl/)
[![GitHub package.json dependency version (subfolder of monorepo)](https://img.shields.io/github/package-json/dependency-version/release-argus/argus/react?filename=web%2Fui%2Freact-app%2Fpackage.json)](https://reactjs.org/)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/release-argus/argus/Test?label=Tests)](https://github.com/release-argus/Argus/actions/workflows/test.yml)
[![Codecov](https://img.shields.io/codecov/c/github/release-argus/argus)](https://app.codecov.io/gh/release-argus/Argus)


[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/release-argus/argus/Binary%20Build?label=Binary%20Build)](https://github.com/release-argus/Argus/actions/workflows/build-binary.yml)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/release-argus/argus)](https://github.com/release-argus/Argus/releases)
[![GitHub all releases](https://img.shields.io/github/downloads/release-argus/argus/total)](https://github.com/release-argus/Argus/releases)
[![GitHub release (latest by SemVer)](https://img.shields.io/github/downloads/release-argus/argus/latest/total)](https://github.com/release-argus/Argus/releases/latest)

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/release-argus/argus/Docker%20Build?label=Docker%20Build)](https://github.com/release-argus/Argus/actions/workflows/build-docker.yml)
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/releaseargus/argus?sort=semver)](https://hub.docker.com/r/releaseargus/argus/tags)
[![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/releaseargus/argus?sort=semver)](https://hub.docker.com/r/releaseargus/argus/tags)
[![Docker Pulls](https://img.shields.io/docker/pulls/releaseargus/argus)](https://hub.docker.com/r/releaseargus/argus)

Argus will query websites at a user defined interval for new software releases and then trigger Gotify/Slack/Other notification(s) and/or WebHook(s) when one has been found.
For example, you could set it to monitor the Argus repo ([release-argus/argus](https://github.com/release-argus/Argus)). This will query the [GitHub API](https://api.github.com/repos/release-argus/argus/releases) and track the "tag_name" variable. When this variable changes from what it was on a previous query, a GitHub-style WebHook could be sent that triggers  something (like AWX) to update Argus on your server.

##### Table of Contents

- [Argus](#argus)
  - [Demo](#demo)
  - [Command-line arguments](#command-line-arguments)
  - [Building from source](#building-from-source)
    - [Prereqs](#prereqs)
    - [Go changes](#go-changes)
    - [React changes](#react-changes)
  - [Getting started](#config-formatting)
    - [Config formatting](#getting-started)

## Demo

A demo of Argus can be seen on our website [here](https://argus.io/demo).

## Command-line arguments

```bash
$ argus -h
Usage of /usr/local/bin/argus:
  -config.check
        Print the fully-parsed config.
  -config.file string
        Argus configuration file path. (default "config.yml")
  -log.level string
        ERROR, WARN, INFO, VERBOSE or DEBUG (default "INFO")
  -log.timestamps
        Enable timestamps in CLI output.
  -test.notify string
        Put the name of the Notify service to send a test message.
  -test.service string
        Put the name of the Service to test the version query.
  -web.cert-file string
        HTTPS certificate file path.
  -web.listen-host string
        IP address to listen on for UI, API, and telemetry. (default "0.0.0.0")
  -web.listen-port string
        Port to listen on for UI, API, and telemetry. (default "8080")
  -web.pkey-file string
        HTTPS private key file path.
  -web.route-prefix string
        Prefix for web endpoints (default "/")
```

## Building from source

#### Prereqs

The backend of Argus is built with [Go](https://go.dev/) and the frontend with [React](https://reactjs.org/). The React frontend is built and then [embedded](https://pkg.go.dev/embed) into the Go binary so that those web files can be served.
- [Go 1.19+](https://go.dev/dl/)
- [NodeJS 16](https://nodejs.org/en/download/)

#### Go changes

To see the changes you've made by modifying any of the `.go` files, you must recompile Argus. You could recompile the whole app with a `make build`, but this will also recompile the React components. To save time (and CPU power), you can use the existing React static and recompile just the Go part by running `make go-build`. (Running this in the root dir will produce the `argus` binary)

#### React changes

To see the changes after modifying anything in `web/ui/react-app`, you must recompile both the Go backend as well as the React frontend. This can be done by running `make build`. (Running this in the root dir will produce the `argus` binary)

## Getting started

To get started with Argus, simply download the binary from the [releases page](https://github.com/release-argus/Argus/releases), and setup the config for that binary.

For further help, check out the [Getting Started](https://release-argus.io/docs/getting-started/) page on our website.

#### Config formatting

The config can be broken down into 5 key areas. ([Further help](https://release-argus.io/docs/config/))
- [defaults](https://release-argus.io/docs/config/defaults/) - This is broken down into areas with defaults for [services](https://release-argus.io/docs/config/defaults/#service-portion), [notify](https://release-argus.io/docs/config/defaults/#notify-portion) and [webhooks](https://release-argus.io/docs/config/defaults/#webhook-portion).
- [settings](https://release-argus.io/docs/config/settings/) - Settings for the Argus server.
- [service](https://release-argus.io/docs/config/service/) - A dictionary mapping of all the services to monitor as well as what to notify when a new release is found.
- [notify](https://release-argus.io/docs/config/notify/) - A dictionary mapping of targets for Notify messages.
- [webhook](https://release-argus.io/docs/config/webhook/) - A dictionary mapping of targets for WebHooks.
