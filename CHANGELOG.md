# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.4.0](https://github.com/release-argus/Argus/compare/0.3.2...0.4.0) (2022-05-14)


### Features

* rename to argus ([f45388a](https://github.com/release-argus/Argus/commits/f45388a8a75a21e3c7f95bd842727739df7771e9))

### [0.3.2](https://github.com/release-argus/Argus/compare/0.3.1...0.3.2) (2022-05-06)


### Bug Fixes

* text when update available for deployed_service ([#28](https://github.com/release-argus/Argus/issues/28)) ([2f7143b](https://github.com/release-argus/Argus/commits/2f7143b5ab7f9047268638ed1b91e24dec1e5dc3))

### 0.3.1 (2022-05-03)


### Bug Fixes

* **refactor:** switch `Service.Status.*` away from pointers ([b579759](https://github.com/release-argus/Argus/commits/b57975925ba1df44889b62cad1bd4ce7eaa9a165))

## 0.3.0 (2022-05-03)


### Features

* `/api/v1/version` ([#19](https://github.com/release-argus/Argus/issues/19)) ([acd33ba](https://github.com/release-argus/Argus/commits/acd33ba08a9c8d3c18c390cdf6d48bb57f3628e1))
* **github-action**: Docker builds - [DockerHub](https://hub.docker.com/r/releaseargus/argus), [GHCR](https://github.com/release-argus/Argus/pkgs/container/argus), [Quay](https://quay.io/repository/argus-io/argus) ([#20](https://github.com/release-argus/Argus/pull/20)) ([6cd458c](https://github.com/release-argus/Argus/commit/6cd458c36a44856c82daefff36781032673b6272))
* **release**: compress web files in github action builds ([3d396c0](https://github.com/release-argus/Argus/commit/3d396c0e75528c45ebf924f83a0fc26c7788d234))


### Bug Fixes

* omit some undefined vars from `config.yml` ([70d049d](https://github.com/release-argus/Argus/commits/70d049daa37247e4199f1e85d3c4b7e30b356454))

### [0.2.1](https://github.com/release-argus/Argus/compare/0.2.0...0.2.1) (2022-05-01)


### Bug Fixes

* **ui:** `manifest.json` icon srcs ([563d062](https://github.com/release-argus/Argus/commits/563d062f4196391bab37900f5b1ced420052c487))

## [0.2.0](https://github.com/release-argus/Argus/compare/0.1.3...0.2.0) (2022-05-01)


### Features

* **query:** semver check `deployed_version` ([594442d](https://github.com/release-argus/Argus/commits/594442d7e2a19ce1c8d6d791b68ced61c0dad7ed))
* **ui:** add icon to show if monitoring deployed service ([07eb53a](https://github.com/release-argus/Argus/commits/07eb53afe404d4cf2280bbe6fa108007b2bb7fb8))


### Bug Fixes

* **config:** don't require global defaults for notifiers/webhook ([4b9751f](https://github.com/release-argus/Argus/commits/4b9751f6c1d09d1dae5d7600c56f4a9b0df964bf))
* correct gotify/webhook invalid config prints ([105c342](https://github.com/release-argus/Argus/commits/105c3423bd7692fe55a28d35eebd30a933f26989))
* default `current_version` to `latest_version` if undefined ([50b0408](https://github.com/release-argus/Argus/commits/50b04088f93ba53ed04bdda5d0e6c2e7bdbfe468))

### [0.1.3](https://github.com/release-argus/Argus/compare/0.1.2...0.1.3) (2022-04-29)


### Bug Fixes

* **config:** switch `listen-address` to `listen-host` (`listen_address` -> `listen_host` as well) ([478ff0e](https://github.com/release-argus/Argus/commit/478ff0ead7260b5577df504f38c50fa01acc4d09))

### [0.1.2](https://github.com/release-argus/Argus/compare/0.1.1...0.1.2) (2022-04-29)


### Bug Fixes

* `UpdateLatestApproved` - handle `nil` webhooks ([1e43e73](https://github.com/release-argus/Argus/commits/1e43e73b4b12aa59c526974a3537d7286e64c17e))

### [0.1.1](https://github.com/release-argus/Argus/compare/0.1.0...0.1.1) (2022-04-25)


### Bug Fixes

* **ui:** icons could overflow ([#17](https://github.com/release-argus/Argus/issues/17)) ([e3897f4](https://github.com/release-argus/Argus/commits/e3897f419a59395d5c292d0c4e34dfa83e641f11))

## [0.1.0](https://github.com/release-argus/Argus/compare/0.0.1...0.1.0) (2022-04-23)


### Features

* **query:** support for retrieving `current_version` from a deployed service ([#12](https://github.com/release-argus/Argus/issues/12)) ([3ebf785](https://github.com/release-argus/Argus/commits/3ebf785f28595d6a57c4f297155cc4c26d9fe94b))

## [0.0.1](https://github.com/release-argus/Argus/compare/0.0.0...0.0.1) (2022-04-23)


### Bug Fixes

* **query:** sort github tag_names when semver ([#11](https://github.com/release-argus/Argus/issues/11)) ([c350c90](https://github.com/release-argus/Argus/commits/c350c90ad67d4a69912671a59200ed610e8b7ab2))

## 0.0.0 (2022-04-21)


### Features

* **initial-release:** [Release-Notifier](https://github.com/JosephKav/Release-Notifier) with WebUI. Has support for Gotify/Slack/WebHooks on new software releases being found. ([f349ede](https://github.com/release-argus/Argus/commit/f349edee99ef54c0f4057abdfb0955b63ee7ce5b))
