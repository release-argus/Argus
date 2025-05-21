# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.27.0](https://github.com/release-argus/Argus/compare/0.26.2...0.27.0) (2025-05-21)


### Features

* `/api/v1/version` ([#19](https://github.com/release-argus/Argus/issues/19)) ([acd33ba](https://github.com/release-argus/Argus/commit/acd33ba08a9c8d3c18c390cdf6d48bb57f3628e1))
* **api:** add /api/v1/counts endpoint - app stats ([8364209](https://github.com/release-argus/Argus/commit/836420996d5e906bde612b7eabedde28e426a708)), closes [#436](https://github.com/release-argus/Argus/issues/436)
* **approvals:** add keyboard shortcuts for search and clear actions ([9336dfb](https://github.com/release-argus/Argus/commit/9336dfbd833c554b934c1576d917ca01d7c4ce85))
* combine `regex` with `regex_submatch` ([47f5e72](https://github.com/release-argus/Argus/commit/47f5e72012ba52a60d1ae17b01345bef91f20572))
* **command:** apply version var templating to args ([5bf93b7](https://github.com/release-argus/Argus/commit/5bf93b7930978a4c65fe6fb3aaf57f4b9f0b7d56))
* **config,web:** option to disable routes ([#384](https://github.com/release-argus/Argus/issues/384)) ([c2d1fdc](https://github.com/release-argus/Argus/commit/c2d1fdcaa7a8eda44f6b5d333ce5a3c83adda6d6))
* **config:** `active` var to disable a service ([#88](https://github.com/release-argus/Argus/issues/88)) ([af756f4](https://github.com/release-argus/Argus/commit/af756f458cfbe9f379a96e338376d20c1702976c))
* **config:** `comment` var for services ([#90](https://github.com/release-argus/Argus/issues/90)) ([a6b68eb](https://github.com/release-argus/Argus/commit/a6b68eb72495f8ff450feba1b9cb2f47fe1523db))
* **config:** hash `web.basic_auth.(username|password)` locally ([#358](https://github.com/release-argus/Argus/issues/358)) ([1ed247b](https://github.com/release-argus/Argus/commit/1ed247be35ffc173a280abcb6526905da32eadc9))
* **config:** load environment variables from a `.env` file ([3006932](https://github.com/release-argus/Argus/commit/300693285a4b2ae7defde15fedda9c09655c88d3))
* **config:** more env var support ([#367](https://github.com/release-argus/Argus/issues/367)) ([87d63d6](https://github.com/release-argus/Argus/commit/87d63d680bc4c5815452d479a0fe3dcd6ae2ab04))
* **config:** set defaults with env vars ([4eca113](https://github.com/release-argus/Argus/commit/4eca113157b907fc121c55e2800e41678a2aa33f))
* **config:** settings from env vars ([ee4d6bf](https://github.com/release-argus/Argus/commit/ee4d6bfbff6a3846ff4ec1613b6b589a9743a3a4))
* **config:** style reformat ([#135](https://github.com/release-argus/Argus/issues/135)) ([5830e1a](https://github.com/release-argus/Argus/commit/5830e1ad5d829863fc30368b396ce4e671bfd278))
* **config:** support env vars in some config vars ([c5d5532](https://github.com/release-argus/Argus/commit/c5d5532e2ed17ea59ab49b6e5195c1caa096c34e))
* **deployed_version:** `target_header` to get version from resp header ([b99363f](https://github.com/release-argus/Argus/commit/b99363f1fd8ad94578ce58fc5a1dd66c1ced4e2f))
* **deployed_version:** add `manual` type ([3f402f6](https://github.com/release-argus/Argus/commit/3f402f62c87a60f8c8596f8bab277cb9121f1376))
* **deployed_version:** regex templating ([#347](https://github.com/release-argus/Argus/issues/347)) ([249379b](https://github.com/release-argus/Argus/commit/249379bbe250f88e2d69cf885c4926af26ed191a))
* **deployed_version:** support arrays in json filter ([0e0bdb1](https://github.com/release-argus/Argus/commit/0e0bdb10ac1a44fb56f2ea0d2a31b9e96a5be62d)), closes [#292](https://github.com/release-argus/Argus/issues/292)
* **deployed_version:** support for post requests ([#398](https://github.com/release-argus/Argus/issues/398)) ([9504d7d](https://github.com/release-argus/Argus/commit/9504d7de5fd1429ed86b805dc9c9e3de299766af)), closes [#397](https://github.com/release-argus/Argus/issues/397)
* **docker:** add `/api/v1/healthcheck` endpoint and binary ([589287a](https://github.com/release-argus/Argus/commit/589287a1d4ac27ee4f1d4931af59327e0246fc41))
* **docker:** add curl ([#394](https://github.com/release-argus/Argus/issues/394)) ([1bc9041](https://github.com/release-argus/Argus/commit/1bc9041e4c7634468d39549b7e857a17e185de70))
* improve help on toolbar ([e15aed9](https://github.com/release-argus/Argus/commit/e15aed9516ea890cdbe964094d10f64c84a4fd91))
* **latest_version:** add pagination to GitHub release fetching ([03726ad](https://github.com/release-argus/Argus/commit/03726adf9af72c9668ab615471895c776d5ce794))
* **latest_version:** regex templating ([1d374a0](https://github.com/release-argus/Argus/commit/1d374a01f920224d9cc22c2c5a8465d1a552d824))
* **log:** print tags that don't match the regex ([1bdfee3](https://github.com/release-argus/Argus/commit/1bdfee3031d349d2244dfc5d11ab271056e140df))
* **metrics:** add, latest_version_is_deployed ([104d50d](https://github.com/release-argus/Argus/commit/104d50d43a0625a802ad3f4c9fac16e8da942d1f)), closes [#293](https://github.com/release-argus/Argus/issues/293)
* **notify:** jinja templating on params ([#101](https://github.com/release-argus/Argus/issues/101)) ([b1bf54e](https://github.com/release-argus/Argus/commit/b1bf54e52ce3c04a8812ca003c80814c243aa3fc))
* **notify:** shoutrrr 'generic' webhook ([b01ada0](https://github.com/release-argus/Argus/commit/b01ada0a9cb77591ee7ec8f2f0f20cc3c1983eb3)), closes [#271](https://github.com/release-argus/Argus/issues/271)
* open external links in a new tab ([26df3b3](https://github.com/release-argus/Argus/commit/26df3b3cc8f7de1d2c1a2c7e23bf9aa56f511a33))
* **query:** semver check `deployed_version` ([594442d](https://github.com/release-argus/Argus/commit/594442d7e2a19ce1c8d6d791b68ced61c0dad7ed))
* **query:** support for retrieving `current_version` from a deployed service ([#12](https://github.com/release-argus/Argus/issues/12)) ([3ebf785](https://github.com/release-argus/Argus/commit/3ebf785f28595d6a57c4f297155cc4c26d9fe94b))
* rename to argus ([f45388a](https://github.com/release-argus/Argus/commit/f45388a8a75a21e3c7f95bd842727739df7771e9))
* **require:** command to accept/reject versions ([#144](https://github.com/release-argus/Argus/issues/144)) ([799c9fc](https://github.com/release-argus/Argus/commit/799c9fc1986af1da8ae537a2837a5735f9d94293))
* **require:** docker tags ([#143](https://github.com/release-argus/Argus/issues/143)) ([b575030](https://github.com/release-argus/Argus/commit/b575030efc59ac62ff85d0fca753cbc1ec788a08))
* **service:** add /tags fallback to github services ([207e610](https://github.com/release-argus/Argus/commit/207e610793b4660811b51760418a24ed36720052)), closes [#275](https://github.com/release-argus/Argus/issues/275)
* **service:** cache ServiceInfo on Status for faster lookups ([ad5540a](https://github.com/release-argus/Argus/commit/ad5540a6272612f3eee7e24b968a347b5dc787b1))
* **service:** default service notify/command/webhook ([08eb05f](https://github.com/release-argus/Argus/commit/08eb05ff7ecc6ceb235704eb9943a3ea2294b929))
* **service:** github, conditional requests ([#176](https://github.com/release-argus/Argus/issues/176)) ([cccab2b](https://github.com/release-argus/Argus/commit/cccab2bdb3ca0b0bfc5446a816cb07357806746a))
* **service:** support leading v in versions (e.g. v1.2.3) ([bdab68d](https://github.com/release-argus/Argus/commit/bdab68daa9f085ae3b3fd617bd6b10ae52aa7d0f))
* **shoutrrr:** 0.7 support - bark/ntfy ([0fc85a8](https://github.com/release-argus/Argus/commit/0fc85a86ac8e645c1b6e5acdd52a3b7415a1f392)), closes [#235](https://github.com/release-argus/Argus/issues/235)
* sqlite db for state - `argus.db` ([#113](https://github.com/release-argus/Argus/issues/113)) ([4502c6d](https://github.com/release-argus/Argus/commit/4502c6d39c9f545e0a7190c378fc3505b51bbd03))
* support for running os commands as a new release action ([#70](https://github.com/release-argus/Argus/issues/70)) ([c6a9e75](https://github.com/release-argus/Argus/commit/c6a9e75a8f804ece5b0dff65b00d90710edd66fc))
* Support for Telegram ([#63](https://github.com/release-argus/Argus/issues/63)) ([f1b9960](https://github.com/release-argus/Argus/commit/f1b996026e7b6d9f2f2c6ea929df7a3c52e5b944))
* **ui:** add arias to components for improved accessibility ([5f62563](https://github.com/release-argus/Argus/commit/5f6256341fa6be8dd1b0ac43f07687281ea7616c))
* **ui:** add icon to show if monitoring deployed service ([07eb53a](https://github.com/release-argus/Argus/commit/07eb53afe404d4cf2280bbe6fa108007b2bb7fb8))
* **ui:** add links to icon_url/web_url and visible lv/dv links on focus ([49a5b02](https://github.com/release-argus/Argus/commit/49a5b027c05016095a323404e7d725086d436431))
* **ui:** allow resending of webhooks to update deployed version ([#110](https://github.com/release-argus/Argus/issues/110)) ([862cc97](https://github.com/release-argus/Argus/commit/862cc9774f715a6f2575c571b8b640d40ad67e59))
* **ui:** can skip releases for a `deployed_service` without webhooks ([#73](https://github.com/release-argus/Argus/issues/73)) ([cf2a816](https://github.com/release-argus/Argus/commit/cf2a816d22b9daef75dbae464831a4e1040396b2))
* **ui:** disable search functionality for small select components ([1b9cc5a](https://github.com/release-argus/Argus/commit/1b9cc5a60449f6e2fb72a870904ee0c3c1f78589))
* **ui:** icons can be links - `icon_link_to` ([#92](https://github.com/release-argus/Argus/issues/92)) ([8c3a9af](https://github.com/release-argus/Argus/commit/8c3a9af5c38a8ea3ce16fdf1c62ab0a730a16359))
* **ui:** move toolbar args to query params and parse templates with api ([a4e2e87](https://github.com/release-argus/Argus/commit/a4e2e87f6255814a7171bd95888f48b6b4d5334e))
* **ui:** re-arrange services ([faa7fc0](https://github.com/release-argus/Argus/commit/faa7fc0fe75aba7e92e0ddc455cd1b218264d0d4))
* **ui:** service tags ([9a81e62](https://github.com/release-argus/Argus/commit/9a81e6213421f2d8ef10c4e1bd6e35df71b90632))
* **web:** 'test notify' button on create/edit ([#379](https://github.com/release-argus/Argus/issues/379)) ([29a5fe0](https://github.com/release-argus/Argus/commit/29a5fe0e2dc97a6a8ff2bec9c6713f01b64cb1e1))
* **web:** basic auth ([#140](https://github.com/release-argus/Argus/issues/140)) ([85fd943](https://github.com/release-argus/Argus/commit/85fd94338db158a648cf72f045ccb8dd72047bad))
* **web:** clickable links for latest/deployed version url ([#385](https://github.com/release-argus/Argus/issues/385)) ([18ae353](https://github.com/release-argus/Argus/commit/18ae3533cdf67dc201907b4ad6cbdf99c1c6a0d3))
* **web:** custom favicon support ([007db66](https://github.com/release-argus/Argus/commit/007db662f31723aa6c31bf2db8f661f2258ef85e))
* **web:** editable services ([#226](https://github.com/release-argus/Argus/issues/226)) ([b5a1b8a](https://github.com/release-argus/Argus/commit/b5a1b8afaf970881201802634e7238f362442c58))
* **webhook:** add `gitlab` type ([#95](https://github.com/release-argus/Argus/issues/95)) ([5a8ab55](https://github.com/release-argus/Argus/commit/5a8ab551e672f08e6a82fdcbd3e0d3bc8f498c0d))
* **webhook:** apply version var templating to custom headers + url ([e31f51b](https://github.com/release-argus/Argus/commit/e31f51bc9877a40101e74c9cb930f2adf9be564d))
* **webhooks:** add option to allow invalid https certs ([#74](https://github.com/release-argus/Argus/issues/74)) ([e390882](https://github.com/release-argus/Argus/commit/e390882c7a3706b23d80cfddcee31ddcf2322ca0))
* **webhook:** support for custom headers ([#83](https://github.com/release-argus/Argus/issues/83)) ([4cc0a14](https://github.com/release-argus/Argus/commit/4cc0a14cb3ecc441e803b565b82c092b76302c06))
* **web:** lv url-commands, add regex index field ([#274](https://github.com/release-argus/Argus/issues/274)) ([7def08c](https://github.com/release-argus/Argus/commit/7def08c65be748fab603506e8d7e83cf1be1dab8))


### Bug Fixes

* `UpdateLatestApproved` - handle `nil` webhooks ([1e43e73](https://github.com/release-argus/Argus/commit/1e43e73b4b12aa59c526974a3537d7286e64c17e))
* 75aad65c6fb14b771c144c01ae2640e53ec3a22e was checking too early ([fb109d1](https://github.com/release-argus/Argus/commit/fb109d1e792299a9563588eef2c00d900fb9d88b))
* allow negative split `index` ([ab61399](https://github.com/release-argus/Argus/commit/ab613995654a5bbc7d3d7855105c1195bf59aed4))
* **build:** add `package-lock.json` to version bumps ([617b9e2](https://github.com/release-argus/Argus/commit/617b9e2c3e95d263f4147cfb8a6746b7c616c9f3))
* check that `latest_versions` are newer than `deployed_version` (and !=) ([75aad65](https://github.com/release-argus/Argus/commit/75aad65c6fb14b771c144c01ae2640e53ec3a22e))
* **config,web:** disabled routes weren't taking into account `route_prefix` ([79213ca](https://github.com/release-argus/Argus/commit/79213ca322473a5c49cd058d0effc4044e71331c))
* **config:** don't require global defaults for notifiers/webhook ([4b9751f](https://github.com/release-argus/Argus/commit/4b9751f6c1d09d1dae5d7600c56f4a9b0df964bf))
* **config:** handle commented out services ([#455](https://github.com/release-argus/Argus/issues/455)) ([d1b9e8c](https://github.com/release-argus/Argus/commit/d1b9e8cf27591407fc6bf595f50d13eb8ccec3ae))
* **config:** ignore comments when extracting order ([c3d0eed](https://github.com/release-argus/Argus/commit/c3d0eed71572a600d94cc3259b9aa20827587aee))
* **config:** start save handler with config load ([#137](https://github.com/release-argus/Argus/issues/137)) ([0a9d866](https://github.com/release-argus/Argus/commit/0a9d866cdf0c45e0f8b33cec4ccb65029b7d3825))
* **config:** switch `listen-address` to `listen-host` ([478ff0e](https://github.com/release-argus/Argus/commit/478ff0ead7260b5577df504f38c50fa01acc4d09))
* **config:** use both converted and unconverted `active` vars ([#139](https://github.com/release-argus/Argus/issues/139)) ([c01b8a8](https://github.com/release-argus/Argus/commit/c01b8a8a334dca239e9a6ac97ceb5143c6014919))
* correct gotify/webhook invalid config prints ([105c342](https://github.com/release-argus/Argus/commit/105c3423bd7692fe55a28d35eebd30a933f26989))
* **db:** can't insert new rows with an `UPDATE` ([#133](https://github.com/release-argus/Argus/issues/133)) ([934a845](https://github.com/release-argus/Argus/commit/934a845c400f6d6ca44190ffdf949523c1167919))
* **db:** only convert if there are services to convert ([1f0938f](https://github.com/release-argus/Argus/commit/1f0938ff72cc67c3a40fa99c6f595ac533dc83f8))
* **db:** switch to text for versions to keep trailing 0's ([7289ce9](https://github.com/release-argus/Argus/commit/7289ce9f5cb1aae0d658d2ec5c69b77f52daa4fe))
* default `current_version` to `latest_version` if undefined ([50b0408](https://github.com/release-argus/Argus/commit/50b04088f93ba53ed04bdda5d0e6c2e7bdbfe468))
* **deployed_version:** null semVer default comp ([70c0625](https://github.com/release-argus/Argus/commit/70c062522917f865b3ae04ec5bbdebf339ef8ba0))
* **deployed_version:** req 2xx for version queries ([#310](https://github.com/release-argus/Argus/issues/310)) ([c2c70fc](https://github.com/release-argus/Argus/commit/c2c70fc830394a58948a467160559fa2aa91aa53))
* **docker:** add OCI header for GHCR ([#503](https://github.com/release-argus/Argus/issues/503)) ([94bc9db](https://github.com/release-argus/Argus/commit/94bc9dbc5c98868d3129b05e0b9c768ba128e177))
* **docker:** allow periods in image name ([#248](https://github.com/release-argus/Argus/issues/248)) ([502c6a9](https://github.com/release-argus/Argus/commit/502c6a96281ea4299dcfbd86b324da507472bfc3))
* **docker:** don't fail startup if chown fails ([80f2e61](https://github.com/release-argus/Argus/commit/80f2e61fa1526943884da76a9d28e742b8a3351e))
* **docker:** remove hardcoded USER to restore custom UID/GID support ([#520](https://github.com/release-argus/Argus/issues/520)) ([799a5a6](https://github.com/release-argus/Argus/commit/799a5a60f04dd6d29119076e668e100bd54737e6))
* **edit:** properly remove notify defaults that aren't overriden ([a49939d](https://github.com/release-argus/Argus/commit/a49939d52a64d8da50255e524a11290664fe80cf))
* ensure used webhooks have a `type` on startup ([3eda486](https://github.com/release-argus/Argus/commit/3eda486d55e18ef6a8a3d02bd58d251587b12833))
* **errors:** include service name in webhook/notify error messages ([#138](https://github.com/release-argus/Argus/issues/138)) ([a8e303f](https://github.com/release-argus/Argus/commit/a8e303f6613e4de377b802490759ef25690acedb))
* **https:** verify cert/key exist ([a24e194](https://github.com/release-argus/Argus/commit/a24e1948020cadcc888c33a739c2a896213902b0))
* **latest_version:** double check new releases ([9e319dd](https://github.com/release-argus/Argus/commit/9e319dd554166e4ffa847ae3784278de39610f8f))
* **log:** initialise before first use ([#529](https://github.com/release-argus/Argus/issues/529)) ([87d2fcb](https://github.com/release-argus/Argus/commit/87d2fcb0cffb85e4bc7d0203a1dbed59b68de14e))
* **metrics:** `latest_version_is_deployed` was cleared on first query ([10ebd35](https://github.com/release-argus/Argus/commit/10ebd35bc4a1e5172805230653596686be77739d))
* notify wasn't passing params down ([fc7201a](https://github.com/release-argus/Argus/commit/fc7201a6a6e7aed3bce492080a437a2fe78ba64a))
* **notify:** crashed with notify in service, but none in notify global ([66a19fc](https://github.com/release-argus/Argus/commit/66a19fc087b1c4cf8b5659d5ee2d46b86be8d9ba))
* **notify:** don't login to matrix to verify (could run into a rate-limit) ([#198](https://github.com/release-argus/Argus/issues/198)) ([a15437d](https://github.com/release-argus/Argus/commit/a15437d62434300a3cf03aefd67a5ee9eb63e68b))
* **notify:** evaluate environment variables in params ([#580](https://github.com/release-argus/Argus/issues/580)) ([82af40a](https://github.com/release-argus/Argus/commit/82af40a18360a1ff05cdfa60606105be0d87e273))
* **notify:** get access_token `password`s working for matrix ([#196](https://github.com/release-argus/Argus/issues/196)) ([74282d1](https://github.com/release-argus/Argus/commit/74282d1bdf6f6b7fc8179d02abe8ba0c6bc6d43c))
* **notify:** handle `#`s in matrix `params.rooms` ([#197](https://github.com/release-argus/Argus/issues/197)) ([81a2a72](https://github.com/release-argus/Argus/commit/81a2a72dc2405096b846c91daf9d88162511fd87))
* **notify:** missing defaults for shoutrrr type ([4bdbf17](https://github.com/release-argus/Argus/commit/4bdbf172f220fb3706a9aaed66d89d2b46b94501))
* **notify:** support 'fromName' in SMTP messages ([#513](https://github.com/release-argus/Argus/issues/513)) ([e47c332](https://github.com/release-argus/Argus/commit/e47c332de69968cf3826bd5440d57c17c64c06ff))
* **notify:** was missing `shoutrrr:` from the url ([2e269df](https://github.com/release-argus/Argus/commit/2e269df39beb03363155681e57ea0feaf1df2096))
* **notify:** was requiring `email` instead of `smtp`. will use `smtp` now ([369dad8](https://github.com/release-argus/Argus/commit/369dad80a11df79682136efe368f64f9d6acf6a8))
* omit some undefined vars from `config.yml` ([70d049d](https://github.com/release-argus/Argus/commit/70d049daa37247e4199f1e85d3c4b7e30b356454))
* ordering struggled with empty newlines ([9a1b4a1](https://github.com/release-argus/Argus/commit/9a1b4a1a26fe02791e77df4962fc2dce058dcc4f))
* properly convert `regex_submatch` to `regex` ([2fa3942](https://github.com/release-argus/Argus/commit/2fa3942af6678876c273f3c50c52bec6f5b27ffe))
* **query:** sort github tag_names when semver ([#11](https://github.com/release-argus/Argus/issues/11)) ([c350c90](https://github.com/release-argus/Argus/commit/c350c90ad67d4a69912671a59200ed610e8b7ab2))
* **refactor:** switch `Service.Status.*` away from pointers ([b579759](https://github.com/release-argus/Argus/commit/b57975925ba1df44889b62cad1bd4ce7eaa9a165))
* Rename `current_version` to `deployed_version` ([#64](https://github.com/release-argus/Argus/issues/64)) ([0b3b1d5](https://github.com/release-argus/Argus/commit/0b3b1d568cf334765322a488a22d3a41467ae2ff))
* require `index` var for `regex` url_command ([#35](https://github.com/release-argus/Argus/issues/35)) ([bda54f4](https://github.com/release-argus/Argus/commit/bda54f491c8ffa3c63df346145692b8953e08217))
* **require:** use auth for hub/quay Docker lookups ([#147](https://github.com/release-argus/Argus/issues/147)) ([d032dfb](https://github.com/release-argus/Argus/commit/d032dfba72fb9ef1bc5f34861a775c8c95b0efd3))
* save sometimes copied a random blank line ([76da7b2](https://github.com/release-argus/Argus/commit/76da7b21459c59ae7f8eac1a2c43a6e1ff166b07))
* **saving:** empty `service.notify.name`s were removed ([#76](https://github.com/release-argus/Argus/issues/76)) ([a2471b0](https://github.com/release-argus/Argus/commit/a2471b02a0db1ffe0b1e086f51c4d50656b8bff6))
* **service:** close latest/deployed version http connections ([#154](https://github.com/release-argus/Argus/issues/154)) ([739c7d8](https://github.com/release-argus/Argus/commit/739c7d84078a494c655dc85f7d8584fce76d0737))
* **service:** use `Connection` header on HTTP requests ([#156](https://github.com/release-argus/Argus/issues/156)) ([6f46976](https://github.com/release-argus/Argus/commit/6f46976a58fe3d0a776d241d52facd97c9b09338))
* **skips:** only update approved_version if commands/webhooks are with a deployed_version lookup ([6972734](https://github.com/release-argus/Argus/commit/69727342f8e01c052a08c3a73a1b38216f99d83a))
* text when update available for deployed_service ([#28](https://github.com/release-argus/Argus/issues/28)) ([2f7143b](https://github.com/release-argus/Argus/commit/2f7143b5ab7f9047268638ed1b91e24dec1e5dc3))
* **ui:** `manifest.json` icon srcs ([563d062](https://github.com/release-argus/Argus/commit/563d062f4196391bab37900f5b1ced420052c487))
* **ui:** adjust padding logic for full-width columns ([ccab296](https://github.com/release-argus/Argus/commit/ccab2960e5f394b159353b2612ee46dddf092450))
* **ui:** adjust position padding on RegEx url_commands in service edit ([56c2a84](https://github.com/release-argus/Argus/commit/56c2a84b012f86f711c47b18db2cdc452c7e390e))
* **ui:** always give type on new service version refresh ([7f90e17](https://github.com/release-argus/Argus/commit/7f90e178718e83b8854e1896faeb94683a650569)), closes [#524](https://github.com/release-argus/Argus/issues/524)
* **ui:** ensure boolean options are always processed with strToBool ([dd13797](https://github.com/release-argus/Argus/commit/dd137978d9abe39373d191139ab87afb49df654a)), closes [#517](https://github.com/release-argus/Argus/issues/517)
* **ui:** handle null oldObj in deepDiff to prevent refresh failure ([#515](https://github.com/release-argus/Argus/issues/515)) ([757c716](https://github.com/release-argus/Argus/commit/757c716c8aaac3e75812e11a12346223f9995842))
* **ui:** have VersionWithRefresh depend on original data ([#516](https://github.com/release-argus/Argus/issues/516)) ([7d16750](https://github.com/release-argus/Argus/commit/7d1675010a33f277c8148a8ec1147ab066a0afbf))
* **ui:** icons could overflow ([#17](https://github.com/release-argus/Argus/issues/17)) ([e3897f4](https://github.com/release-argus/Argus/commit/e3897f419a59395d5c292d0c4e34dfa83e641f11))
* **ui:** improve deployed_version `manual` type input handling ([704c2cf](https://github.com/release-argus/Argus/commit/704c2cfcc450be606e1cf28fb6288dd41f810d89))
* **ui:** improve version conversion logic for service edit ([3f15f89](https://github.com/release-argus/Argus/commit/3f15f89b697349a4a5341b2680e5aae8324d2868))
* **ui:** only send deployed_version type with other values ([92d9c16](https://github.com/release-argus/Argus/commit/92d9c16b25e16c4e0570a799f39424112aeadbc9)), closes [#539](https://github.com/release-argus/Argus/issues/539)
* **ui:** re-add error display for version fetch in VersionWithRefresh ([4a2fe33](https://github.com/release-argus/Argus/commit/4a2fe3319e4e3efe0009ea2468dbbe57bb025e54))
* **ui:** show error on first version refresh failure ([2e0479f](https://github.com/release-argus/Argus/commit/2e0479ff75fc4b86896bd1c1cd49354a5dba5c3c))
* **ui:** skip wasn't working for services with no commands/webhooks ([#127](https://github.com/release-argus/Argus/issues/127)) ([43ce3bf](https://github.com/release-argus/Argus/commit/43ce3bf4705aa41567ba921a34bbe1518c2553e4))
* **version-refresh:** clear version errors after fetch ([e968847](https://github.com/release-argus/Argus/commit/e96884764e1cbdab9f41f049efaa5b33220a7b66))
* **version-refresh:** omit version from query keys ([ec8b3b2](https://github.com/release-argus/Argus/commit/ec8b3b2f2807029c0cb970eea4c2639d4a5b8d0e))
* **web:** allow skip when command/webhook blocked by delay ([4436d38](https://github.com/release-argus/Argus/commit/4436d3828f1f4e89456ccea97a41b485bd9f1efb))
* **web:** compare previous `semantic_version` state in version refreshes ([3414a24](https://github.com/release-argus/Argus/commit/3414a2432487834c7bc1629448d182b665138eb3)), closes [#279](https://github.com/release-argus/Argus/issues/279)
* **web:** default `Name` to `ID` when creating or editing a service ([58d1284](https://github.com/release-argus/Argus/commit/58d1284cd98713e9dcf28803fe5ce340430574ee))
* **web:** formcheck width on non-webkit browsers ([cf1d681](https://github.com/release-argus/Argus/commit/cf1d68153640dceafa8ece42b8f55a3a223ce2d2))
* **webhook:** use `custom_headers` from main/defaults ([#174](https://github.com/release-argus/Argus/issues/174)) ([0ba4124](https://github.com/release-argus/Argus/commit/0ba412420bed4922978483ca5016f7abd4f89e41))
* **web:** ntfy, default fieldValues for actions ([#456](https://github.com/release-argus/Argus/issues/456)) ([280ebc1](https://github.com/release-argus/Argus/commit/280ebc17731b1db636723586b44cb49392a5c253))
* **web:** render icon/icon_link_to/web_url changes from websocket ([78a5656](https://github.com/release-argus/Argus/commit/78a5656ea7a90398e4083105b9b82a87ac6d74fe))
* **web:** show link for `deployed_version.url` ([bcbd835](https://github.com/release-argus/Argus/commit/bcbd835d0bd94e9323cff728337bca307192964d))
* **web:** use the semantic_versioning bool on query ([#270](https://github.com/release-argus/Argus/issues/270)) ([fd4e2ae](https://github.com/release-argus/Argus/commit/fd4e2ae13ac52fe97bf5afbb1be79ca6344ca819))

## [0.26.2](https://github.com/release-argus/Argus/compare/0.26.1...0.26.2) (2025-05-16)


### Bug Fixes

* **notify:** evaluate environment variables in params ([#580](https://github.com/release-argus/Argus/issues/580)) ([2f80722](https://github.com/release-argus/Argus/commit/2f807222064ee37759ff857256d4a49cb269387d))

## [0.26.1](https://github.com/release-argus/Argus/compare/0.26.0...v0.26.1) (2025-05-15)


### Bug Fixes

* **ui:** show error on first version refresh failure ([2e0479f](https://github.com/release-argus/Argus/commit/2e0479ff75fc4b86896bd1c1cd49354a5dba5c3c))
* **web:** default `Name` to `ID` when creating or editing a service ([58d1284](https://github.com/release-argus/Argus/commit/58d1284cd98713e9dcf28803fe5ce340430574ee))

## [0.26.0](https://github.com/release-argus/Argus/compare/0.25.0...0.26.0) (2025-05-10)


### Features

* **latest_version:** add pagination to GitHub release fetching ([03726ad](https://github.com/release-argus/Argus/commits/03726adf9af72c9668ab615471895c776d5ce794))


### Bug Fixes

* **version-refresh:** clear version errors after fetch ([e968847](https://github.com/release-argus/Argus/commits/e96884764e1cbdab9f41f049efaa5b33220a7b66))
* **version-refresh:** omit version from query keys ([ec8b3b2](https://github.com/release-argus/Argus/commits/ec8b3b2f2807029c0cb970eea4c2639d4a5b8d0e))

## [0.25.0](https://github.com/release-argus/Argus/compare/0.24.0...0.25.0) (2025-04-30)


### Features

* **config:** load environment variables from a `.env` file ([3006932](https://github.com/release-argus/Argus/commits/300693285a4b2ae7defde15fedda9c09655c88d3))
* **service:** cache ServiceInfo on Status for faster lookups ([ad5540a](https://github.com/release-argus/Argus/commits/ad5540a6272612f3eee7e24b968a347b5dc787b1))


### Bug Fixes

* **ui:** adjust padding logic for full-width columns ([ccab296](https://github.com/release-argus/Argus/commits/ccab2960e5f394b159353b2612ee46dddf092450))
* **ui:** adjust position padding on RegEx url_commands in service edit ([56c2a84](https://github.com/release-argus/Argus/commits/56c2a84b012f86f711c47b18db2cdc452c7e390e))
* **ui:** re-add error display for version fetch in VersionWithRefresh ([4a2fe33](https://github.com/release-argus/Argus/commits/4a2fe3319e4e3efe0009ea2468dbbe57bb025e54))

## [0.24.0](https://github.com/release-argus/Argus/compare/0.23.0...0.24.0) (2025-04-08)


### Features

* **ui:** move toolbar args to query params and parse templates with api ([a4e2e87](https://github.com/release-argus/Argus/commits/a4e2e87f6255814a7171bd95888f48b6b4d5334e))
* **ui:** re-arrange services ([faa7fc0](https://github.com/release-argus/Argus/commits/faa7fc0fe75aba7e92e0ddc455cd1b218264d0d4))

## [0.23.0](https://github.com/release-argus/Argus/compare/0.22.0...0.23.0) (2025-03-09)


### Features

* **ui:** add links to icon_url/web_url and visible lv/dv links on focus ([49a5b02](https://github.com/release-argus/Argus/commits/49a5b027c05016095a323404e7d725086d436431))


### Bug Fixes

* **ui:** improve deployed_version `manual` type input handling ([704c2cf](https://github.com/release-argus/Argus/commits/704c2cfcc450be606e1cf28fb6288dd41f810d89))

## [0.22.0](https://github.com/release-argus/Argus/compare/0.21.0...0.22.0) (2025-03-08)


### Features

* **deployed_version:** `target_header` to get version from resp header ([b99363f](https://github.com/release-argus/Argus/commits/b99363f1fd8ad94578ce58fc5a1dd66c1ced4e2f))
* **deployed_version:** add `manual` type ([3f402f6](https://github.com/release-argus/Argus/commits/3f402f62c87a60f8c8596f8bab277cb9121f1376))
* **ui:** disable search functionality for small select components ([1b9cc5a](https://github.com/release-argus/Argus/commits/1b9cc5a60449f6e2fb72a870904ee0c3c1f78589))


### Bug Fixes

* **ui:** only send deployed_version type with other values ([92d9c16](https://github.com/release-argus/Argus/commits/92d9c16b25e16c4e0570a799f39424112aeadbc9)), closes [#539](https://github.com/release-argus/Argus/issues/539)

### [0.21.0](https://github.com/release-argus/Argus/compare/0.20.0...0.21.0) (2025-02-08)


### Features

* **ui:** add arias to components for improved accessibility ([5f62563](https://github.com/release-argus/Argus/commits/5f6256341fa6be8dd1b0ac43f07687281ea7616c))


### Bug Fixes

* **log:** initialise before first use ([#529](https://github.com/release-argus/Argus/issues/529)) ([87d2fcb](https://github.com/release-argus/Argus/commits/87d2fcb0cffb85e4bc7d0203a1dbed59b68de14e)), closes [#528](https://github.com/release-argus/Argus/issues/528)
* **ui:** always give type on new service version refresh ([7f90e17](https://github.com/release-argus/Argus/commits/7f90e178718e83b8854e1896faeb94683a650569)), closes [#524](https://github.com/release-argus/Argus/issues/524)
* **ui:** improve version conversion logic for service edit ([3f15f89](https://github.com/release-argus/Argus/commits/3f15f89b697349a4a5341b2680e5aae8324d2868))

## [0.20.0](https://github.com/release-argus/Argus/compare/0.19.4...0.20.0) (2025-01-26)


### Features

* **ui:** service tags ([9a81e62](https://github.com/release-argus/Argus/commits/9a81e6213421f2d8ef10c4e1bd6e35df71b90632))


### Bug Fixes

* **ui:** ensure boolean options are always processed with strToBool ([dd13797](https://github.com/release-argus/Argus/commits/dd137978d9abe39373d191139ab87afb49df654a)), closes [#517](https://github.com/release-argus/Argus/issues/517)

### [0.19.4](https://github.com/release-argus/Argus/compare/0.19.3...0.19.4) (2025-01-18)


### Bug Fixes

* **docker:** remove hardcoded USER to restore custom UID/GID support ([#520](https://github.com/release-argus/Argus/issues/520)) ([799a5a6](https://github.com/release-argus/Argus/commits/799a5a60f04dd6d29119076e668e100bd54737e6))

### [0.19.3](https://github.com/release-argus/Argus/compare/0.19.2...0.19.3) (2025-01-15)


### Bug Fixes

* **ui:** handle null oldObj in deepDiff to prevent refresh failure ([#515](https://github.com/release-argus/Argus/issues/515)) ([757c716](https://github.com/release-argus/Argus/commits/757c716c8aaac3e75812e11a12346223f9995842))
* **ui:** have VersionWithRefresh depend on original data ([#516](https://github.com/release-argus/Argus/issues/516)) ([7d16750](https://github.com/release-argus/Argus/commits/7d1675010a33f277c8148a8ec1147ab066a0afbf))

### [0.19.2](https://github.com/release-argus/Argus/compare/0.19.1...0.19.2) (2025-01-14)


### Bug Fixes

* **notify:** support 'fromName' in SMTP messages ([#513](https://github.com/release-argus/Argus/issues/513)) ([e47c332](https://github.com/release-argus/Argus/commits/e47c332de69968cf3826bd5440d57c17c64c06ff))

### [0.19.1](https://github.com/release-argus/Argus/compare/0.18.0...0.19.1) (2025-01-14)


### Bug Fixes

* **config:** ignore comments when extracting order ([c3d0eed](https://github.com/release-argus/Argus/commits/c3d0eed71572a600d94cc3259b9aa20827587aee))

## [0.19.0](https://github.com/release-argus/Argus/compare/0.18.0...0.19.0) (2025-01-13)


### Features

* **api:** add `/api/v1/counts` endpoint - app stats ([8364209](https://github.com/release-argus/Argus/commits/836420996d5e906bde612b7eabedde28e426a708)), closes [#436](https://github.com/release-argus/Argus/issues/436)
    * service_count
    * updates_available
    * updates_skipped (still counts as available)
* **approvals:** add keyboard shortcuts for search and clear actions ([9336dfb](https://github.com/release-argus/Argus/commits/9336dfbd833c554b934c1576d917ca01d7c4ce85))
    * `/` to focus on the service filter bar.
    * `esc` whilst focused on that search bar to clear it and blur it.


### Bug Fixes

* **config:** handle commented out services ([#455](https://github.com/release-argus/Argus/issues/455)) ([d1b9e8c](https://github.com/release-argus/Argus/commits/d1b9e8cf27591407fc6bf595f50d13eb8ccec3ae))
* **deployed_version:** null semVer default comp ([70c0625](https://github.com/release-argus/Argus/commits/70c062522917f865b3ae04ec5bbdebf339ef8ba0))
* **docker:** add OCI header for GHCR ([#503](https://github.com/release-argus/Argus/issues/503)) ([94bc9db](https://github.com/release-argus/Argus/commits/94bc9dbc5c98868d3129b05e0b9c768ba128e177))
* **web:** ntfy, default fieldValues for actions ([#456](https://github.com/release-argus/Argus/issues/456)) ([280ebc1](https://github.com/release-argus/Argus/commits/280ebc17731b1db636723586b44cb49392a5c253))
* **web:** render icon/icon_link_to/web_url changes from websocket ([78a5656](https://github.com/release-argus/Argus/commits/78a5656ea7a90398e4083105b9b82a87ac6d74fe))

## [0.18.0](https://github.com/release-argus/Argus/compare/0.17.4...0.18.0) (2024-05-07)


### Features

* **deployed_version:** support for POST requests ([#398](https://github.com/release-argus/Argus/issues/398)) ([9504d7d](https://github.com/release-argus/Argus/commits/9504d7de5fd1429ed86b805dc9c9e3de299766af)), closes [#397](https://github.com/release-argus/Argus/issues/397)


### Bug Fixes

* **edit:** properly remove notify defaults that aren't overridden ([a49939d](https://github.com/release-argus/Argus/commits/a49939d52a64d8da50255e524a11290664fe80cf))

### [0.17.4](https://github.com/release-argus/Argus/compare/0.17.3...0.17.4) (2024-04-27)


### Bug Fixes

* **docker:** don't fail startup if chown fails ([80f2e61](https://github.com/release-argus/Argus/commits/80f2e61fa1526943884da76a9d28e742b8a3351e)), closes [#106](https://github.com/release-argus/Argus/issues/106)

### [0.17.3](https://github.com/release-argus/Argus/compare/0.17.0...0.17.3) (2024-04-23)


### Features

* **docker:** add curl ([#394](https://github.com/release-argus/Argus/issues/394)) ([1bc9041](https://github.com/release-argus/Argus/commits/1bc9041e4c7634468d39549b7e857a17e185de70))


### Bug Fixes

* **config,web:** disabled routes weren't taking into account `route_prefix` ([79213ca](https://github.com/release-argus/Argus/commits/79213ca322473a5c49cd058d0effc4044e71331c))
* **web:** show link for `deployed_version.url` ([bcbd835](https://github.com/release-argus/Argus/commits/bcbd835d0bd94e9323cff728337bca307192964d))

### [0.17.2](https://github.com/release-argus/Argus/compare/0.17.0...0.17.2) (2024-04-15)


### Bug Fixes

* **web:** show link for `deployed_version.url` ([bcbd835](https://github.com/release-argus/Argus/commits/bcbd835d0bd94e9323cff728337bca307192964d))

### [0.17.1](https://github.com/release-argus/Argus/compare/0.17.0...0.17.1) (2024-04-15)


### Bug Fixes

* **config,web:** disabled routes weren't taking into account `route_prefix` ([79213ca](https://github.com/release-argus/Argus/commits/79213ca322473a5c49cd058d0effc4044e71331c))

## [0.17.0](https://github.com/release-argus/Argus/compare/0.16.0...0.17.0) (2024-04-15)


### Features

* **config,web:** option to disable routes ([#384](https://github.com/release-argus/Argus/issues/384)) ([c2d1fdc](https://github.com/release-argus/Argus/commits/c2d1fdcaa7a8eda44f6b5d333ce5a3c83adda6d6))
    * `settings.web.disabled_routes: []`
* **web:** 'test notify' button on create/edit ([#379](https://github.com/release-argus/Argus/issues/379)) ([29a5fe0](https://github.com/release-argus/Argus/commits/29a5fe0e2dc97a6a8ff2bec9c6713f01b64cb1e1))
* **web:** clickable links for latest/deployed version url ([#385](https://github.com/release-argus/Argus/issues/385)) ([18ae353](https://github.com/release-argus/Argus/commits/18ae3533cdf67dc201907b4ad6cbdf99c1c6a0d3))

## [0.16.0](https://github.com/release-argus/Argus/compare/0.15.2...0.16.0) (2024-02-24)

New mascot! Thanks [@rexapex](https://github.com/rexapex)

### Features

* **config:** hash `web.basic_auth.(username|password)` locally ([#358](https://github.com/release-argus/Argus/issues/358)) ([1ed247b](https://github.com/release-argus/Argus/commits/1ed247be35ffc173a280abcb6526905da32eadc9))
* **config:** more env var support ([#367](https://github.com/release-argus/Argus/issues/367)) ([87d63d6](https://github.com/release-argus/Argus/commits/87d63d680bc4c5815452d479a0fe3dcd6ae2ab04))
* **config:** support env vars in some config vars ([#365](https://github.com/release-argus/Argus/pull/365)) ([c5d5532](https://github.com/release-argus/Argus/commits/c5d5532e2ed17ea59ab49b6e5195c1caa096c34e))
* **web:** custom favicon support ([#355](https://github.com/release-argus/Argus/pull/355)) ([007db66](https://github.com/release-argus/Argus/commits/007db662f31723aa6c31bf2db8f661f2258ef85e))

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
