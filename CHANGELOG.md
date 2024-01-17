# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [0.15.2](https://github.com/release-argus/Argus/compare/0.15.1...0.15.2) (2024-01-17)


### Bug Fixes

* **latest_version:** double check new releases ([5610a28](https://github.com/release-argus/Argus/commits/5610a28d6982f3033e041903a3ded75585ac4e74))

### [0.15.1](https://github.com/release-argus/Argus/compare/0.15.0...0.15.1) (2024-01-13)


### Bug Fixes

* **web:** formcheck width on non-webkit browsers ([cf1d681](https://github.com/release-argus/Argus/commits/cf1d68153640dceafa8ece42b8f55a3a223ce2d2))

## [0.15.0](https://github.com/release-argus/Argus/compare/0.14.0...0.15.0) (2024-01-09)


### Features

* **deployed_version:** regex templating ([#347](https://github.com/release-argus/Argus/issues/347)) ([249379b](https://github.com/release-argus/Argus/commits/249379bbe250f88e2d69cf885c4926af26ed191a))

## [0.14.0](https://github.com/release-argus/Argus/compare/0.13.3...0.14.0) (2024-01-06)


### Features

* **latest_version:** regex templating ([1d374a0](https://github.com/release-argus/Argus/commits/1d374a01f920224d9cc22c2c5a8465d1a552d824))

### [0.13.3](https://github.com/release-argus/Argus/compare/0.13.2...0.13.3) (2023-09-12)


### Features

* **notify:** shoutrrr 'generic' webhook ([b01ada0](https://github.com/release-argus/Argus/commits/b01ada0a9cb77591ee7ec8f2f0f20cc3c1983eb3)), closes [#271](https://github.com/release-argus/Argus/issues/271)


### Bug Fixes

* **deployed_version:** req 2xx for version queries ([#310](https://github.com/release-argus/Argus/issues/310)) ([c2c70fc](https://github.com/release-argus/Argus/commits/c2c70fc830394a58948a467160559fa2aa91aa53))

### [0.13.2](https://github.com/release-argus/Argus/compare/0.13.1...0.13.2) (2023-07-19)


### Bug Fixes

* **metrics:** `latest_version_is_deployed` was cleared on first query ([b8e6093](https://github.com/release-argus/Argus/commits/b8e6093a4b35e6d4d8617149e1412d1a0718da94))
    * metrics are reset, but this one wasn't re-created in that reset

    * could also sometimes have metrics for inactive services

    * `*_version_query_result_last` metric wasn't deleted with the service

### [0.13.1](https://github.com/release-argus/Argus/compare/0.13.0...0.13.1) (2023-07-19)


### Features

* **deployed_version:** support arrays in json filter ([1f1e1d0](https://github.com/release-argus/Argus/commits/1f1e1d0624cf3d822d08bbb7420fae016eeef644)), closes [#292](https://github.com/release-argus/Argus/issues/292)
    * e.g. `foo[0].version`
* **metrics:** add, `latest_version_is_deployed` ([8ba1074](https://github.com/release-argus/Argus/commits/8ba107496f6a4e3732de5a681147d0138c29255f)), closes [#293](https://github.com/release-argus/Argus/issues/293)
    * merged `ack_waiting` into this
    * 0=no, 1=yes, 2=approved, 3=skipped

## [0.13.0](https://github.com/release-argus/Argus/compare/0.12.1...0.13.0) (2023-07-11)


### Features

* **service:** add /tags fallback to github services ([207e610](https://github.com/release-argus/Argus/commits/207e610793b4660811b51760418a24ed36720052)), closes [#275](https://github.com/release-argus/Argus/issues/275)
* **service:** default service notify/command/webhook ([08eb05f](https://github.com/release-argus/Argus/commits/08eb05ff7ecc6ceb235704eb9943a3ea2294b929))
* **service:** support leading v in versions (e.g. v1.2.3) ([bdab68d](https://github.com/release-argus/Argus/commits/bdab68daa9f085ae3b3fd617bd6b10ae52aa7d0f))
* **web:** lv url-commands, add regex index field ([#274](https://github.com/release-argus/Argus/issues/274)) ([7def08c](https://github.com/release-argus/Argus/commits/7def08c65be748fab603506e8d7e83cf1be1dab8))


### Bug Fixes

* **db:** switch to text for versions to keep trailing 0's ([7289ce9](https://github.com/release-argus/Argus/commits/7289ce9f5cb1aae0d658d2ec5c69b77f52daa4fe))
* **notify:** missing defaults for shoutrrr type ([4bdbf17](https://github.com/release-argus/Argus/commits/4bdbf172f220fb3706a9aaed66d89d2b46b94501))
* **web:** allow skip when command/webhook blocked by delay ([4436d38](https://github.com/release-argus/Argus/commits/4436d3828f1f4e89456ccea97a41b485bd9f1efb))
* **web:** compare previous `semantic_version` state in version refreshes ([3414a24](https://github.com/release-argus/Argus/commits/3414a2432487834c7bc1629448d182b665138eb3)), closes [#279](https://github.com/release-argus/Argus/issues/279)

### [0.12.1](https://github.com/release-argus/Argus/compare/0.12.0...0.12.1) (2023-06-19)


### Bug Fixes

* **web:** use the semantic_versioning bool on query ([#270](https://github.com/release-argus/Argus/issues/270)) ([fd4e2ae](https://github.com/release-argus/Argus/commits/fd4e2ae13ac52fe97bf5afbb1be79ca6344ca819))

### [0.12.0](https://github.com/release-argus/Argus/compare/0.11.1...0.12.0) (2023-06-01)


### Features

* **config:** set defaults with env vars ([4eca113](https://github.com/release-argus/Argus/commits/4eca113157b907fc121c55e2800e41678a2aa33f))
    * anything under defaults, e.g. **ARGUS_SERVICE_LATEST_VERSION_ACCESS_TOKEN** for `defaults.service.latest_version.access_token` in the YAML
    * anything under settings, e.g. **ARGUS_DATA_DATABASE_FILE** for `settings.data.database_file` in the YAML
* **shoutrrr:** 0.7 support - bark/ntfy ([0fc85a8](https://github.com/release-argus/Argus/commits/0fc85a86ac8e645c1b6e5acdd52a3b7415a1f392)), closes [#235](https://github.com/release-argus/Argus/issues/235)
* **web:** editable services ([#226](https://github.com/release-argus/Argus/issues/226)) ([b5a1b8a](https://github.com/release-argus/Argus/commits/b5a1b8afaf970881201802634e7238f362442c58))


### Bug Fixes

* **docker:** allow periods in image name ([#248](https://github.com/release-argus/Argus/issues/248)) ([502c6a9](https://github.com/release-argus/Argus/commits/502c6a96281ea4299dcfbd86b324da507472bfc3))

### [0.11.1](https://github.com/release-argus/Argus/compare/0.11.0...0.11.1) (2023-01-13)


### Bug Fixes

* **notify:** don't login to matrix to verify (could run into a rate-limit) ([#198](https://github.com/release-argus/Argus/issues/198)) ([a15437d](https://github.com/release-argus/Argus/commits/a15437d62434300a3cf03aefd67a5ee9eb63e68b))
* **notify:** get access_token `password`s working for matrix ([#196](https://github.com/release-argus/Argus/issues/196)) ([74282d1](https://github.com/release-argus/Argus/commits/74282d1bdf6f6b7fc8179d02abe8ba0c6bc6d43c))
* **notify:** handle `#`s in matrix `params.rooms` ([#197](https://github.com/release-argus/Argus/issues/197)) ([81a2a72](https://github.com/release-argus/Argus/commits/81a2a72dc2405096b846c91daf9d88162511fd87))

## [0.11.0](https://github.com/release-argus/Argus/compare/0.10.3...0.11.0) (2022-11-18)


### Features

* **docker:** add `/api/v1/healthcheck` endpoint and binary ([589287a](https://github.com/release-argus/Argus/commits/589287a1d4ac27ee4f1d4931af59327e0246fc41))
* **service:** github, conditional requests ([#176](https://github.com/release-argus/Argus/issues/176)) ([cccab2b](https://github.com/release-argus/Argus/commits/a1e9deac47899029828cbedd6fc52eeb2198d758))


### Bug Fixes

* **webhook:** use `custom_headers` from main/defaults ([#174](https://github.com/release-argus/Argus/issues/174)) ([0ba4124](https://github.com/release-argus/Argus/commits/0ba412420bed4922978483ca5016f7abd4f89e41))

### [0.10.3](https://github.com/release-argus/Argus/compare/0.10.2...0.10.3) (2022-09-15)


### Bug Fixes

* **service:** use `Connection` header on HTTP requests ([#156](https://github.com/release-argus/Argus/issues/156)) ([6f46976](https://github.com/release-argus/Argus/commits/6f46976a58fe3d0a776d241d52facd97c9b09338))

### [0.10.2](https://github.com/release-argus/Argus/compare/0.10.1...0.10.2) (2022-09-06)


### Bug Fixes

* **service:** close latest/deployed version http connections ([#154](https://github.com/release-argus/Argus/issues/154)) ([739c7d8](https://github.com/release-argus/Argus/commits/739c7d84078a494c655dc85f7d8584fce76d0737))

### [0.10.1](https://github.com/release-argus/Argus/compare/0.10.0...0.10.1) (2022-08-24)


### Bug Fixes

* **require:** use auth for hub/quay Docker lookups ([#147](https://github.com/release-argus/Argus/issues/147)) ([d032dfb](https://github.com/release-argus/Argus/commits/d032dfba72fb9ef1bc5f34861a775c8c95b0efd3))

## [0.10.0](https://github.com/release-argus/Argus/compare/0.9.0...0.10.0) (2022-08-21)


### Features

* **require:** command to accept/reject versions ([#144](https://github.com/release-argus/Argus/issues/144)) ([799c9fc](https://github.com/release-argus/Argus/commits/799c9fc1986af1da8ae537a2837a5735f9d94293))
* **require:** docker tags ([#143](https://github.com/release-argus/Argus/issues/143)) ([b575030](https://github.com/release-argus/Argus/commits/b575030efc59ac62ff85d0fca753cbc1ec788a08))

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
