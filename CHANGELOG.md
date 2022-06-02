# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.5.0](https://github.com/release-argus/Argus/compare/0.4.1...0.5.0) (2022-06-02)


### Features

* combine `regex` with `regex_submatch` ([47f5e72](https://github.com/release-argus/Argus/commits/47f5e72012ba52a60d1ae17b01345bef91f20572))
* improve help on toolbar ([e15aed9](https://github.com/release-argus/Argus/commits/e15aed9516ea890cdbe964094d10f64c84a4fd91))
* open external links in a new tab ([26df3b3](https://github.com/release-argus/Argus/commits/26df3b3cc8f7de1d2c1a2c7e23bf9aa56f511a33))
* support for running os commands as a new release action ([#70](https://github.com/release-argus/Argus/issues/70)) ([c6a9e75](https://github.com/release-argus/Argus/commits/c6a9e75a8f804ece5b0dff65b00d90710edd66fc))
* Support for Telegram and lots more ([#63](https://github.com/release-argus/Argus/issues/63)) ([f1b9960](https://github.com/release-argus/Argus/commits/f1b996026e7b6d9f2f2c6ea929df7a3c52e5b944))
* **ui:** can skip releases for a `deployed_service` without webhooks ([#73](https://github.com/release-argus/Argus/issues/73)) ([cf2a816](https://github.com/release-argus/Argus/commits/cf2a816d22b9daef75dbae464831a4e1040396b2))
* **webhooks:** add option to allow invalid https certs ([#74](https://github.com/release-argus/Argus/issues/74)) ([e390882](https://github.com/release-argus/Argus/commits/e390882c7a3706b23d80cfddcee31ddcf2322ca0))


### Bug Fixes

* allow negative split `index` ([ab61399](https://github.com/release-argus/Argus/commits/ab613995654a5bbc7d3d7855105c1195bf59aed4))
* check that `latest_version`s are newer than `deployed_version` (and !=) ([75aad65](https://github.com/release-argus/Argus/commits/75aad65c6fb14b771c144c01ae2640e53ec3a22e))
* ensure used webhooks have a `type` on startup ([3eda486](https://github.com/release-argus/Argus/commits/3eda486d55e18ef6a8a3d02bd58d251587b12833)), closes [/github.com/release-argus/Argus/issues/71#issuecomment-1139456015](https://github.com/release-argus//github.com/release-argus/Argus/issues/71/issues/issuecomment-1139456015)
* ordering struggled with empty newlines ([9a1b4a1](https://github.com/release-argus/Argus/commits/9a1b4a1a26fe02791e77df4962fc2dce058dcc4f))
* Rename `current_version` to `deployed_version` ([#64](https://github.com/release-argus/Argus/issues/64)) ([0b3b1d5](https://github.com/release-argus/Argus/commits/0b3b1d568cf334765322a488a22d3a41467ae2ff))
* save sometimes copied a random blank line ([76da7b2](https://github.com/release-argus/Argus/commits/76da7b21459c59ae7f8eac1a2c43a6e1ff166b07))

### [0.4.1](https://github.com/release-argus/Argus/compare/0.4.0...0.4.1) (2022-05-16)


### Bug Fixes

* require `index` var for `regex` url_command ([#35](https://github.com/release-argus/Argus/issues/35)) ([bda54f4](https://github.com/release-argus/Argus/commits/bda54f491c8ffa3c63df346145692b8953e08217))

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
