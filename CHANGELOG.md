# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.9.0](https://github.com/release-argus/Argus/compare/0.8.2...0.9.0) (2022-08-10)


### Features

* **config:** non-breaking style reformat ([#135](https://github.com/release-argus/Argus/issues/135)) ([5830e1a](https://github.com/release-argus/Argus/commits/5830e1ad5d829863fc30368b396ce4e671bfd278))
* **web:** basic auth ([#140](https://github.com/release-argus/Argus/issues/140)) ([85fd943](https://github.com/release-argus/Argus/commits/85fd94338db158a648cf72f045ccb8dd72047bad))


### Bug Fixes

* **errors:** include service name in webhook/notify error messages ([#138](https://github.com/release-argus/Argus/issues/138)) ([a8e303f](https://github.com/release-argus/Argus/commits/a8e303f6613e4de377b802490759ef25690acedb))

### [0.8.2](https://github.com/release-argus/Argus/compare/0.8.1...0.8.2) (2022-07-30)


### Bug Fixes

* **db:** can't insert new rows with an `UPDATE` ([#133](https://github.com/release-argus/Argus/issues/133)) ([934a845](https://github.com/release-argus/Argus/commits/934a845c400f6d6ca44190ffdf949523c1167919))

### [0.8.1](https://github.com/release-argus/Argus/compare/0.8.0...0.8.1) (2022-07-22)


### Bug Fixes

* **db:** only convert if there are services to convert ([1f0938f](https://github.com/release-argus/Argus/commits/1f0938ff72cc67c3a40fa99c6f595ac533dc83f8))
* **ui:** skip wasn't working for services with no commands/webhooks ([#127](https://github.com/release-argus/Argus/issues/127)) ([43ce3bf](https://github.com/release-argus/Argus/commits/43ce3bf4705aa41567ba921a34bbe1518c2553e4))

## [0.8.0](https://github.com/release-argus/Argus/compare/0.7.0...0.8.0) (2022-07-16)


### Features

* **notify:** jinja templating on params ([#101](https://github.com/release-argus/Argus/issues/101)) ([b1bf54e](https://github.com/release-argus/Argus/commits/b1bf54e52ce3c04a8812ca003c80814c243aa3fc))
* sqlite db for state - `data/argus.db` ([#113](https://github.com/release-argus/Argus/issues/113)) ([41294f3](https://github.com/release-argus/Argus/commits/41294f326b66cf49fdbf836667fecd3cc12e3296))
* **ui:** allow resending of webhooks to update deployed version ([#110](https://github.com/release-argus/Argus/issues/110)) ([862cc97](https://github.com/release-argus/Argus/commits/862cc9774f715a6f2575c571b8b640d40ad67e59))


### Bug Fixes

* **notify:** crashed with notify in service, but none in notify global ([66a19fc](https://github.com/release-argus/Argus/commits/66a19fc087b1c4cf8b5659d5ee2d46b86be8d9ba))

## [0.7.0](https://github.com/release-argus/Argus/compare/0.6.0...0.7.0) (2022-06-22)


### Features

* **command:** apply version var templating to args ([5bf93b7](https://github.com/release-argus/Argus/commits/5bf93b7930978a4c65fe6fb3aaf57f4b9f0b7d56))
* **config:** `active` var to disable a service ([#88](https://github.com/release-argus/argus/issues/88)) ([af756f4](https://github.com/release-argus/Argus/commits/af756f458cfbe9f379a96e338376d20c1702976c))
* **config:** `comment` var for services ([#90](https://github.com/release-argus/argus/issues/90)) ([a6b68eb](https://github.com/release-argus/Argus/commits/a6b68eb72495f8ff450feba1b9cb2f47fe1523db))
* **ui:** icons can be links - `icon_link_to` ([#92](https://github.com/release-argus/argus/issues/92)) ([8c3a9af](https://github.com/release-argus/Argus/commits/8c3a9af5c38a8ea3ce16fdf1c62ab0a730a16359))
* **webhook:** add `gitlab` type ([#95](https://github.com/release-argus/argus/issues/95)) ([5a8ab55](https://github.com/release-argus/Argus/commits/5a8ab551e672f08e6a82fdcbd3e0d3bc8f498c0d))
* **webhook:** apply version var templating to custom headers + url ([e31f51b](https://github.com/release-argus/Argus/commits/e31f51bc9877a40101e74c9cb930f2adf9be564d))


### Bug Fixes

* **https:** verify cert/key exist ([a24e194](https://github.com/release-argus/Argus/commits/a24e1948020cadcc888c33a739c2a896213902b0)), closes [/github.com/release-argus/Argus/issues/84#issuecomment-1150016743](https://github.com/release-argus/Argus/issues/84/#issuecomment-1150016743)

## [0.6.0](https://github.com/release-argus/Argus/compare/0.5.1...0.6.0) (2022-06-04)


### Features

* **webhook:** support for custom headers ([#83](https://github.com/release-argus/Argus/issues/83)) ([4cc0a14](https://github.com/release-argus/Argus/commits/4cc0a14cb3ecc441e803b565b82c092b76302c06))

### [0.5.1](https://github.com/release-argus/Argus/compare/0.5.0...0.5.1) (2022-06-02)


### Bug Fixes

* **skips:** only update approved_version if commands/webhooks are with a deployed_version lookup ([6972734](https://github.com/release-argus/Argus/commits/69727342f8e01c052a08c3a73a1b38216f99d83a)), closes [/github.com/release-argus/Argus/issues/62#issuecomment-1144963060](https://github.com/release-argus/Argus/issues/62/#issuecomment-1144963060)

## [0.5.0](https://github.com/release-argus/Argus/compare/0.4.1...0.5.0) (2022-06-02)


### Features

* combine `regex` with `regex_submatch` ([47f5e72](https://github.com/release-argus/Argus/commits/47f5e72012ba52a60d1ae17b01345bef91f20572))
* improve help on toolbar ([e15aed9](https://github.com/release-argus/Argus/commits/e15aed9516ea890cdbe964094d10f64c84a4fd91))
* open external links in a new tab ([26df3b3](https://github.com/release-argus/Argus/commits/26df3b3cc8f7de1d2c1a2c7e23bf9aa56f511a33))
* support for running os commands as a new release action ([#70](https://github.com/release-argus/Argus/issues/70)) ([c6a9e75](https://github.com/release-argus/Argus/commits/c6a9e75a8f804ece5b0dff65b00d90710edd66fc))
* support for telegram and lots more ([#63](https://github.com/release-argus/Argus/issues/63)) ([f1b9960](https://github.com/release-argus/Argus/commits/f1b996026e7b6d9f2f2c6ea929df7a3c52e5b944))t
* **ui:** can skip releases for a `deployed_service` without webhooks ([#73](https://github.com/release-argus/Argus/issues/73)) ([cf2a816](https://github.com/release-argus/Argus/commits/cf2a816d22b9daef75dbae464831a4e1040396b2))
* **webhooks:** add option to allow invalid https certs ([#74](https://github.com/release-argus/Argus/issues/74)) ([e390882](https://github.com/release-argus/Argus/commits/e390882c7a3706b23d80cfddcee31ddcf2322ca0))


### Bug Fixes

* allow negative split `index` ([ab61399](https://github.com/release-argus/Argus/commits/ab613995654a5bbc7d3d7855105c1195bf59aed4))
* check that `latest_version`s are newer than `deployed_version` (and !=) ([75aad65](https://github.com/release-argus/Argus/commits/75aad65c6fb14b771c144c01ae2640e53ec3a22e))
* ensure used webhooks have a `type` on startup ([3eda486](https://github.com/release-argus/Argus/commits/3eda486d55e18ef6a8a3d02bd58d251587b12833)), closes [/github.com/release-argus/Argus/issues/71#issuecomment-1139456015](https://github.com/release-argus/Argus/issues/71/#issuecomment-1139456015)
* ordering struggled with empty newlines ([9a1b4a1](https://github.com/release-argus/Argus/commits/9a1b4a1a26fe02791e77df4962fc2dce058dcc4f))
* rename `current_version` to `deployed_version` ([#64](https://github.com/release-argus/Argus/issues/64)) ([0b3b1d5](https://github.com/release-argus/Argus/commits/0b3b1d568cf334765322a488a22d3a41467ae2ff))
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
