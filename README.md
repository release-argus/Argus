<h1 align="center" style="border-bottom: none">
    <a href="//release-argus.io" target="_blank"><img alt="Argus" src="/web/ui/static/favicon.svg" height=128></a><br>Argus
</h1>

<div align="center">

  Keeping an eye on releases.

  [![GitHub](https://img.shields.io/github/license/release-argus/argus)](https://github.com/release-argus/Argus/blob/master/LICENSE)
  [![Go Report Card](https://goreportcard.com/badge/github.com/release-argus/Argus)](https://goreportcard.com/report/github.com/release-argus/Argus)
  [![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/release-argus/argus?filename=go.mod)](https://go.dev/dl/)
  [![GitHub package.json dependency version (subfolder of monorepo)](https://img.shields.io/github/package-json/dependency-version/release-argus/argus/react?filename=web%2Fui%2Freact-app%2Fpackage.json)](https://reactjs.org/)
  [![Codecov](https://img.shields.io/codecov/c/github/release-argus/argus)](https://app.codecov.io/gh/release-argus/Argus)
  <br>
  [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/release-argus/Argus/build-binary.yml)](https://github.com/release-argus/Argus/actions/workflows/build-binary.yml)
  [![GitHub release (latest by date)](https://img.shields.io/github/v/release/release-argus/argus)](https://github.com/release-argus/Argus/releases)
  [![GitHub all releases](https://img.shields.io/github/downloads/release-argus/argus/total)](https://github.com/release-argus/Argus/releases)
  [![GitHub release (latest by SemVer)](https://img.shields.io/github/downloads/release-argus/argus/latest/total)](https://github.com/release-argus/Argus/releases/latest)
  <br>
  [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/release-argus/Argus/build-docker.yml)](https://github.com/release-argus/Argus/actions/workflows/build-docker.yml)
  [![Docker Image Version (latest semver)](https://img.shields.io/docker/v/releaseargus/argus?sort=semver)](https://hub.docker.com/r/releaseargus/argus/tags)
  [![Docker Image Size (latest semver)](https://img.shields.io/docker/image-size/releaseargus/argus?sort=semver)](https://hub.docker.com/r/releaseargus/argus/tags)
  [![Docker Pulls](https://img.shields.io/docker/pulls/releaseargus/argus)](https://hub.docker.com/r/releaseargus/argus)

</div>

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

A demo of Argus can be seen on our website [here](https://release-argus.io/demo).

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
  -web.basic-auth.password string
        Password for basic auth
  -web.basic-auth.username string
        Username for basic auth
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
- [Go 1.22+](https://go.dev/dl/)
- [NodeJS 20](https://nodejs.org/en/download/)

#### Go changes

To see the changes you've made by modifying any of the `.go` files, you must compile Argus. Run `make build` the first time to ensure the web components are available locally. Amy future builds that don't need the web-ui to be rebuilt can be done with `make go-build` (faster than `make build`). (Running either of these in the root dir will produce an `argus` binary)

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
