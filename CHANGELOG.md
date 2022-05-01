# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.2.0](https://github.com/hymenaios-io/Hymenaios/compare/0.1.3...0.2.0) (2022-05-01)


### Features

* **query:** semver check `deployed_version` ([594442d](https://github.com/hymenaios-io/Hymenaios/commits/594442d7e2a19ce1c8d6d791b68ced61c0dad7ed))
* **ui:** add icon to show if monitoring deployed service ([07eb53a](https://github.com/hymenaios-io/Hymenaios/commits/07eb53afe404d4cf2280bbe6fa108007b2bb7fb8))


### Bug Fixes

* **config:** don't require global defaults for notifiers/webhook ([4b9751f](https://github.com/hymenaios-io/Hymenaios/commits/4b9751f6c1d09d1dae5d7600c56f4a9b0df964bf))
* correct gotify/webhook invalid config prints ([105c342](https://github.com/hymenaios-io/Hymenaios/commits/105c3423bd7692fe55a28d35eebd30a933f26989))
* default `current_version` to `latest_version` if undefined ([50b0408](https://github.com/hymenaios-io/Hymenaios/commits/50b04088f93ba53ed04bdda5d0e6c2e7bdbfe468))

### [0.1.3](https://github.com/hymenaios-io/Hymenaios/compare/0.1.2...0.1.3) (2022-04-29)


### Bug Fixes

* **config:** switch `listen-address` to `listen-host` (`listen_address` -> `listen_host` as well) ([478ff0e](https://github.com/hymenaios-io/Hymenaios/commit/478ff0ead7260b5577df504f38c50fa01acc4d09))

### [0.1.2](https://github.com/hymenaios-io/Hymenaios/compare/0.1.1...0.1.2) (2022-04-29)


### Bug Fixes

* `UpdateLatestApproved` - handle `nil` webhooks ([1e43e73](https://github.com/hymenaios-io/Hymenaios/commits/1e43e73b4b12aa59c526974a3537d7286e64c17e))

### [0.1.1](https://github.com/hymenaios-io/Hymenaios/compare/0.1.0...0.1.1) (2022-04-25)


### Bug Fixes

* **ui:** icons could overflow ([#17](https://github.com/hymenaios-io/Hymenaios/issues/17)) ([e3897f4](https://github.com/hymenaios-io/Hymenaios/commits/e3897f419a59395d5c292d0c4e34dfa83e641f11))

## [0.1.0](https://github.com/hymenaios-io/Hymenaios/compare/0.0.1...0.1.0) (2022-04-23)


### Features

* **query:** support for retrieving `current_version` from a deployed service ([#12](https://github.com/hymenaios-io/Hymenaios/issues/12)) ([3ebf785](https://github.com/hymenaios-io/Hymenaios/commits/3ebf785f28595d6a57c4f297155cc4c26d9fe94b))

## [0.0.1](https://github.com/hymenaios-io/Hymenaios/compare/0.0.0...0.0.1) (2022-04-23)


### Bug Fixes

* **query:** sort github tag_names when semver ([#11](https://github.com/hymenaios-io/hymenaios/issues/11)) ([c350c90](https://github.com/hymenaios-io/Hymenaios/commits/c350c90ad67d4a69912671a59200ed610e8b7ab2))

## 0.0.0 (2022-04-21)


### Features

* **initial-release:** [Release-Notifier](https://github.com/JosephKav/Release-Notifier) with WebUI. Has support for Gotify/Slack/WebHooks on new software releases being found. ([f349ede](https://github.com/hymenaios-io/Hymenaios/commit/f349edee99ef54c0f4057abdfb0955b63ee7ce5b))
