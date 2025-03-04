## [3.0.1-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v3.0.0...v3.0.1-dev.1) (2025-03-04)

### :bug: Fixes

* helm staged files were not read ([49be5c0](https://github.com/JHOFER-Cloud/helm-ci/commit/49be5c0a0b7c1cce6a1037748278615c438fbf53))

## [3.0.0](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.4...v3.0.0) (2025-03-02)

### ⚠ BREAKING CHANGES

* {values_path}/{stage}/manifest.yml instead of
{values_path}/manifest.yml
* YOU NEED TO MIGRATE TO LIVE/DEV_DOMAINS AND SELECT THE
CORRECT TEMPLATE FOR YOUR CHART (or leave default)

### :sparkles: Features

* add mock test ([2179ddd](https://github.com/JHOFER-Cloud/helm-ci/commit/2179ddde67e57c56b87e7d6d90b27608ebbddc6b))
* add tests ([d5fa427](https://github.com/JHOFER-Cloud/helm-ci/commit/d5fa4277985ce3e02ecc068637d16da2d7dec24d))
* allow using multiple domains by using templates ([90a7034](https://github.com/JHOFER-Cloud/helm-ci/commit/90a703466a8c463a365a27ab31cfca6c6c7b86d5))
* allow yaml and yml ([44fcad4](https://github.com/JHOFER-Cloud/helm-ci/commit/44fcad428c7b31fbf8de72a45b58a8a272529593))
* manifest uses now staged folder layout ([1ee9caa](https://github.com/JHOFER-Cloud/helm-ci/commit/1ee9caaff5457fb3f97413eb765fc382ab947b1a))

### :bug: Fixes

* embedded templates ([9d35c2a](https://github.com/JHOFER-Cloud/helm-ci/commit/9d35c2a51142222937cced6d8800eed68c908b31))
* embedded templates ([961ec01](https://github.com/JHOFER-Cloud/helm-ci/commit/961ec0173194a466152aff3c4232fe4cab14796d))
* replace/set metadata.namespace for custom deployments ([4adfa0a](https://github.com/JHOFER-Cloud/helm-ci/commit/4adfa0aad193d41d81f95dd94c6d959e69ec0a19))

### :memo: Documentation

* remove outdated stuff ([6419745](https://github.com/JHOFER-Cloud/helm-ci/commit/64197457d69d58bde1041cecb23b5bc1d9b21ef4))
* update readme ([9d80211](https://github.com/JHOFER-Cloud/helm-ci/commit/9d802119b79841657a0b88f9093d10b68f401446))
* update readme ([71ac0e5](https://github.com/JHOFER-Cloud/helm-ci/commit/71ac0e519578b6cc51f0b203878dd6f2282ca31f))
* update readme ([992efb2](https://github.com/JHOFER-Cloud/helm-ci/commit/992efb2af8ead032975b5bf1ebd168e006d973f6))
* update tests section ([887f92d](https://github.com/JHOFER-Cloud/helm-ci/commit/887f92d643938207ef6c80d3729816a3da7dedda))

### :zap: Refactor

* split script for easier maintanability and readability ([ec4ca6b](https://github.com/JHOFER-Cloud/helm-ci/commit/ec4ca6bb960129cf5cbd4a39508c73a5ccd06cd3))

### :repeat: CI

* add more commit types to semantic release ([4f47d2c](https://github.com/JHOFER-Cloud/helm-ci/commit/4f47d2c2f88e66e32940626077902381f5a5877b))
* fix semantic release ([c83890a](https://github.com/JHOFER-Cloud/helm-ci/commit/c83890a65a3dbf7db5c24d6e937398df02b2ef5f))
* fix typo in semantic release dependency ([7855fd2](https://github.com/JHOFER-Cloud/helm-ci/commit/7855fd2559753667e2e25c1c730783d75fa639d8))

### :repeat: Chore

* cleanup some code analysis warnings/recommendations ([84bbaf7](https://github.com/JHOFER-Cloud/helm-ci/commit/84bbaf795648127f66ebc2dc55d304365b8930d1))
* improve DEBUG mode ([e1b3b8d](https://github.com/JHOFER-Cloud/helm-ci/commit/e1b3b8ddd65514215e4cfcc0b241c73c89712003))
* **release:** 2.1.0-dev.1 [skip ci] ([20aeadb](https://github.com/JHOFER-Cloud/helm-ci/commit/20aeadb54e12177f8c47d508beca5a3964364e81))
* **release:** 2.1.0-dev.2 [skip ci] ([919f900](https://github.com/JHOFER-Cloud/helm-ci/commit/919f90043ec701bed21c013bf645ea226c948ae6))
* **release:** 3.0.0-dev.1 [skip ci] ([34f1686](https://github.com/JHOFER-Cloud/helm-ci/commit/34f1686a16e784595a0db794eb6643f1f9895b17))
* **release:** 3.0.0-dev.2 [skip ci] ([418d7d2](https://github.com/JHOFER-Cloud/helm-ci/commit/418d7d21cd103a69494b9946a5d099c8df918825))
* **release:** 3.0.0-dev.3 [skip ci] ([4eb1597](https://github.com/JHOFER-Cloud/helm-ci/commit/4eb1597f1bfc01a4008dfe1375458905efa2d10e))
* **release:** 3.0.0-dev.4 [skip ci] ([472b05d](https://github.com/JHOFER-Cloud/helm-ci/commit/472b05df4b1c90515d139166dc1622b172eb9bf3))
* update tests ([67a4523](https://github.com/JHOFER-Cloud/helm-ci/commit/67a4523066113bead45d0e7d0f7dce86dce64134))

## [3.0.0-dev.4](https://github.com/JHOFER-Cloud/helm-ci/compare/v3.0.0-dev.3...v3.0.0-dev.4) (2025-03-02)

### :bug: Fixes

* embedded templates ([9d35c2a](https://github.com/JHOFER-Cloud/helm-ci/commit/9d35c2a51142222937cced6d8800eed68c908b31))

## [3.0.0-dev.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v3.0.0-dev.2...v3.0.0-dev.3) (2025-03-02)

### :bug: Fixes

* embedded templates ([15cd8ba](https://github.com/JHOFER-Cloud/helm-ci/commit/15cd8bac373fdff56cb6c6e89995d0219f855ae4))
* embedded templates ([961ec01](https://github.com/JHOFER-Cloud/helm-ci/commit/961ec0173194a466152aff3c4232fe4cab14796d))

## [3.0.0-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v3.0.0-dev.1...v3.0.0-dev.2) (2025-03-02)

### ⚠ BREAKING CHANGES

* {values_path}/{stage}/manifest.yml instead of
{values_path}/manifest.yml

### :sparkles: Features

* allow yaml and yml ([44fcad4](https://github.com/JHOFER-Cloud/helm-ci/commit/44fcad428c7b31fbf8de72a45b58a8a272529593))
* manifest uses now staged folder layout ([1ee9caa](https://github.com/JHOFER-Cloud/helm-ci/commit/1ee9caaff5457fb3f97413eb765fc382ab947b1a))

### :bug: Fixes

* replace/set metadata.namespace for custom deployments ([4adfa0a](https://github.com/JHOFER-Cloud/helm-ci/commit/4adfa0aad193d41d81f95dd94c6d959e69ec0a19))

## [3.0.0-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.1.0-dev.2...v3.0.0-dev.1) (2025-03-02)

### ⚠ BREAKING CHANGES

* YOU NEED TO MIGRATE TO LIVE/DEV_DOMAINS AND SELECT THE
CORRECT TEMPLATE FOR YOUR CHART (or leave default)

### :sparkles: Features

* allow using multiple domains by using templates ([90a7034](https://github.com/JHOFER-Cloud/helm-ci/commit/90a703466a8c463a365a27ab31cfca6c6c7b86d5))

### :memo: Documentation

* remove outdated stuff ([6419745](https://github.com/JHOFER-Cloud/helm-ci/commit/64197457d69d58bde1041cecb23b5bc1d9b21ef4))
* update tests section ([887f92d](https://github.com/JHOFER-Cloud/helm-ci/commit/887f92d643938207ef6c80d3729816a3da7dedda))

### :repeat: Chore

* cleanup some code analysis warnings/recommendations ([84bbaf7](https://github.com/JHOFER-Cloud/helm-ci/commit/84bbaf795648127f66ebc2dc55d304365b8930d1))
* improve DEBUG mode ([e1b3b8d](https://github.com/JHOFER-Cloud/helm-ci/commit/e1b3b8ddd65514215e4cfcc0b241c73c89712003))
* update tests ([67a4523](https://github.com/JHOFER-Cloud/helm-ci/commit/67a4523066113bead45d0e7d0f7dce86dce64134))

## [2.1.0-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.1.0-dev.1...v2.1.0-dev.2) (2025-03-01)

### :repeat: CI

* add more commit types to semantic release ([4f47d2c](https://github.com/JHOFER-Cloud/helm-ci/commit/4f47d2c2f88e66e32940626077902381f5a5877b))
* fix semantic release ([c83890a](https://github.com/JHOFER-Cloud/helm-ci/commit/c83890a65a3dbf7db5c24d6e937398df02b2ef5f))
* fix typo in semantic release dependency ([7855fd2](https://github.com/JHOFER-Cloud/helm-ci/commit/7855fd2559753667e2e25c1c730783d75fa639d8))

# [2.1.0-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.4...v2.1.0-dev.1) (2025-03-01)


### Features

* add mock test ([2179ddd](https://github.com/JHOFER-Cloud/helm-ci/commit/2179ddde67e57c56b87e7d6d90b27608ebbddc6b))
* add tests ([d5fa427](https://github.com/JHOFER-Cloud/helm-ci/commit/d5fa4277985ce3e02ecc068637d16da2d7dec24d))

## [2.0.4](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.3...v2.0.4) (2025-02-15)


### Bug Fixes

* manifest deployment doesnt create naemspace if it doesnt exist ([d520e18](https://github.com/JHOFER-Cloud/helm-ci/commit/d520e18a20e0acb678602050a1c0c40ffdf2f9ad))
* oci registries not working ([922410f](https://github.com/JHOFER-Cloud/helm-ci/commit/922410f2b83d1580d6a3df5c0169a406936a55ae))
* remove OCI login ([8a55131](https://github.com/JHOFER-Cloud/helm-ci/commit/8a5513166d7922547a14113a3d741c57d37ae093))

## [2.0.4-dev.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.4-dev.2...v2.0.4-dev.3) (2025-02-15)


### Bug Fixes

* remove OCI login ([8a55131](https://github.com/JHOFER-Cloud/helm-ci/commit/8a5513166d7922547a14113a3d741c57d37ae093))

## [2.0.4-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.4-dev.1...v2.0.4-dev.2) (2025-02-15)


### Bug Fixes

* manifest deployment doesnt create naemspace if it doesnt exist ([d520e18](https://github.com/JHOFER-Cloud/helm-ci/commit/d520e18a20e0acb678602050a1c0c40ffdf2f9ad))

## [2.0.4-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.3...v2.0.4-dev.1) (2025-02-15)


### Bug Fixes

* oci registries not working ([922410f](https://github.com/JHOFER-Cloud/helm-ci/commit/922410f2b83d1580d6a3df5c0169a406936a55ae))

## [2.0.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.2...v2.0.3) (2025-02-15)


### Bug Fixes

* customDeployment showing secret when diff is unchanged ([3dbddba](https://github.com/JHOFER-Cloud/helm-ci/commit/3dbddbaa7222e63e09baeeff8cd085fa4abb96b2))
* double encoding of manifest secrets ([007fbfb](https://github.com/JHOFER-Cloud/helm-ci/commit/007fbfbe57f8e2a979d68e6548abb800b9b52811))

## [2.0.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1...v2.0.2) (2025-02-15)


### Bug Fixes

* requires buildx ([049c6c5](https://github.com/JHOFER-Cloud/helm-ci/commit/049c6c5163f720503d65a30459b53ad876bd2a36))

## [2.0.1-dev.8](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.7...v2.0.1-dev.8) (2025-02-15)


### Bug Fixes

* requires buildx ([049c6c5](https://github.com/JHOFER-Cloud/helm-ci/commit/049c6c5163f720503d65a30459b53ad876bd2a36))

## [2.0.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.0...v2.0.1) (2025-02-15)


### Bug Fixes

* CA certificate missing in docker container ([5d08405](https://github.com/JHOFER-Cloud/helm-ci/commit/5d084057b2596e3954f42206e595aa2c666dc16a))
* **ci-image:** ca cert not working ([9de11aa](https://github.com/JHOFER-Cloud/helm-ci/commit/9de11aa0b02653d55a47541e2a6ea88648e61429))
* **ci-image:** issue was traefik replica didnt have access to acme storage ([712132c](https://github.com/JHOFER-Cloud/helm-ci/commit/712132cbb712c185b809bd53ae3755656901831a))
* **ci-image:** missing openssl package ([c503ccf](https://github.com/JHOFER-Cloud/helm-ci/commit/c503ccfdd2d55e0fb147c2e578d45da7415b403e))
* **ci-image:** now ca should work ([12dbb1f](https://github.com/JHOFER-Cloud/helm-ci/commit/12dbb1f7d9054b5e2c76d13e4586ac5a128da2c5))
* **ci-image:** overcomplicated it... ([e3f49fc](https://github.com/JHOFER-Cloud/helm-ci/commit/e3f49fc2c294497c760a3d83928bc2720d3c51f5))
* default vault kvVersion (2) ([4438d5c](https://github.com/JHOFER-Cloud/helm-ci/commit/4438d5cd0b8a010a975747a7b274002e78b4ee28))
* pki not reachable ([60c9749](https://github.com/JHOFER-Cloud/helm-ci/commit/60c97498704fee83879511c2771a235e7f134f01))

## [2.0.1-dev.7](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.6...v2.0.1-dev.7) (2025-02-14)


### Bug Fixes

* **ci-image:** issue was traefik replica didnt have access to acme storage ([712132c](https://github.com/JHOFER-Cloud/helm-ci/commit/712132cbb712c185b809bd53ae3755656901831a))

## [2.0.1-dev.6](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.5...v2.0.1-dev.6) (2025-02-14)


### Bug Fixes

* **ci-image:** overcomplicated it... ([e3f49fc](https://github.com/JHOFER-Cloud/helm-ci/commit/e3f49fc2c294497c760a3d83928bc2720d3c51f5))

## [2.0.1-dev.5](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.4...v2.0.1-dev.5) (2025-02-14)


### Bug Fixes

* **ci-image:** now ca should work ([12dbb1f](https://github.com/JHOFER-Cloud/helm-ci/commit/12dbb1f7d9054b5e2c76d13e4586ac5a128da2c5))

## [2.0.1-dev.4](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.3...v2.0.1-dev.4) (2025-02-14)


### Bug Fixes

* **ci-image:** ca cert not working ([9de11aa](https://github.com/JHOFER-Cloud/helm-ci/commit/9de11aa0b02653d55a47541e2a6ea88648e61429))

## [2.0.1-dev.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.2...v2.0.1-dev.3) (2025-02-14)


### Bug Fixes

* **ci-image:** missing openssl package ([c503ccf](https://github.com/JHOFER-Cloud/helm-ci/commit/c503ccfdd2d55e0fb147c2e578d45da7415b403e))

## [2.0.1-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.1-dev.1...v2.0.1-dev.2) (2025-02-14)


### Bug Fixes

* pki not reachable ([60c9749](https://github.com/JHOFER-Cloud/helm-ci/commit/60c97498704fee83879511c2771a235e7f134f01))

## [2.0.1-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v2.0.0...v2.0.1-dev.1) (2025-02-14)


### Bug Fixes

* CA certificate missing in docker container ([5d08405](https://github.com/JHOFER-Cloud/helm-ci/commit/5d084057b2596e3954f42206e595aa2c666dc16a))
* default vault kvVersion (2) ([4438d5c](https://github.com/JHOFER-Cloud/helm-ci/commit/4438d5cd0b8a010a975747a7b274002e78b4ee28))

# [2.0.0](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.3...v2.0.0) (2025-02-14)


### Bug Fixes

* old container image was used ([bc5a8ef](https://github.com/JHOFER-Cloud/helm-ci/commit/bc5a8efc0cb9dffef3f4f1dd99fdc1e48077e646))
* old container image was used ([b94760f](https://github.com/JHOFER-Cloud/helm-ci/commit/b94760f3df34b8d23183166436c66300775b949f))


### BREAKING CHANGES

* helm-ci_image_tag input needs to be set (at best to the
same version as the workflow input, renovate is your friend ;))

## [1.0.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.2...v1.0.3) (2025-02-14)


### Bug Fixes

* chart name not working ([61ef887](https://github.com/JHOFER-Cloud/helm-ci/commit/61ef887ac7e1124cb34a4376fc1ec0a2f7326ce2))

## [1.0.3-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.2...v1.0.3-dev.1) (2025-02-14)


### Bug Fixes

* chart name not working ([61ef887](https://github.com/JHOFER-Cloud/helm-ci/commit/61ef887ac7e1124cb34a4376fc1ec0a2f7326ce2))

## [1.0.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.1...v1.0.2) (2025-02-14)


### Bug Fixes

* semantic_release ([73fcc51](https://github.com/JHOFER-Cloud/helm-ci/commit/73fcc51ce7baeaa67369a07b64a6a4a44a690393))

## [1.0.2-dev.3](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.2-dev.2...v1.0.2-dev.3) (2025-02-14)


### Bug Fixes

* semantic_release ([73fcc51](https://github.com/JHOFER-Cloud/helm-ci/commit/73fcc51ce7baeaa67369a07b64a6a4a44a690393))

## [1.0.2-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.2-dev.1...v1.0.2-dev.2) (2025-02-14)


### Reverts

* Revert "fix: semantic_release" ([0d02e9b](https://github.com/JHOFER-Cloud/helm-ci/commit/0d02e9b958810418ea1d382824696af60caf3c8c))

## [1.0.2-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.1...v1.0.2-dev.1) (2025-02-14)


### Bug Fixes

* semantic_release ([16afaca](https://github.com/JHOFER-Cloud/helm-ci/commit/16afaca506fea9a4d4fcfe50c635783c861ed136))

## [1.0.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.0...v1.0.1) (2025-02-13)


### Bug Fixes

* cant comment on pull request ([5c7f4f3](https://github.com/JHOFER-Cloud/helm-ci/commit/5c7f4f3c7bcc7cacc07351caa8118f82df0bc6f0))
* deployments failing for normal helm-deployments ([#36](https://github.com/JHOFER-Cloud/helm-ci/issues/36)) ([d1cf28c](https://github.com/JHOFER-Cloud/helm-ci/commit/d1cf28c7e8737ad66d2a37f23a9ef6ea8a5e84ce))

## [1.0.1-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.1-dev.1...v1.0.1-dev.2) (2025-02-13)


### Bug Fixes

* deployments failing for normal helm-deployments ([#36](https://github.com/JHOFER-Cloud/helm-ci/issues/36)) ([d1cf28c](https://github.com/JHOFER-Cloud/helm-ci/commit/d1cf28c7e8737ad66d2a37f23a9ef6ea8a5e84ce))

## [1.0.1-dev.1](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.0...v1.0.1-dev.1) (2025-02-13)


### Bug Fixes

* cant comment on pull request ([5c7f4f3](https://github.com/JHOFER-Cloud/helm-ci/commit/5c7f4f3c7bcc7cacc07351caa8118f82df0bc6f0))

# 1.0.0 (2025-02-10)


### Bug Fixes

=====
* ca mount ([a98a7b0](https://github.com/JHOFER-Cloud/helm-ci/commit/a98a7b00d7878e7ba7241e041fff7aad7538526b))
* chart name wrong ([30ab037](https://github.com/JHOFER-Cloud/helm-ci/commit/30ab037260ac7c6a8a5831bbd5f84050f28d9578))
* **ci:** correct tag evaluation logic in Docker workflow ([17e39ba](https://github.com/JHOFER-Cloud/helm-ci/commit/17e39babf48c5fbf77637d306acaf789d29f3d23))
* custom manifest deployment ([d69e441](https://github.com/JHOFER-Cloud/helm-ci/commit/d69e4411c830e674c7ce6689d27b2494f438e962))
* deploy call ([97559af](https://github.com/JHOFER-Cloud/helm-ci/commit/97559af8ad8aea945bcc66df757983edb9ab0490))
* deployment not triggering ([1bdcaec](https://github.com/JHOFER-Cloud/helm-ci/commit/1bdcaecd8f14bb391792c492d53582f86a3e63ae))
* diff ([1712534](https://github.com/JHOFER-Cloud/helm-ci/commit/1712534e5891dca3396224c8760415c461c02703))
* forgot GitHub* in config struct ([#22](https://github.com/JHOFER-Cloud/helm-ci/issues/22)) ([669a371](https://github.com/JHOFER-Cloud/helm-ci/commit/669a371c549f5b2855c58538a9fe5e1c7a22167b))
* multiline secrets ([f6fc7bb](https://github.com/JHOFER-Cloud/helm-ci/commit/f6fc7bb0909fc417d19a24dc0d811d39728bbec7))
* pr/dev not containg cert and set values ([dcc1409](https://github.com/JHOFER-Cloud/helm-ci/commit/dcc1409f499e78da311b54c9bc2b036afa11bcb6))
* root ca volume ([89f9f35](https://github.com/JHOFER-Cloud/helm-ci/commit/89f9f35281d937ae8aa3cdaceb4f491bf312944b))
* semantic-release issue write permissions ([b288e3a](https://github.com/JHOFER-Cloud/helm-ci/commit/b288e3a4b8334523a797efbb425baf3cb646a7e1))
* setup root ca ([679b045](https://github.com/JHOFER-Cloud/helm-ci/commit/679b0452a87e365d041e69f3c246575fccdfb378))
* values files ([e0cd8df](https://github.com/JHOFER-Cloud/helm-ci/commit/e0cd8df553d156f1a9a6d7bc56504a961e83eb02))
* vault ingress setting not working (FIX THIS BETTER) ([e9911b6](https://github.com/JHOFER-Cloud/helm-ci/commit/e9911b6e6d6676d41227b30acf012692b48ce92b))


### Features

* add diff ([138145d](https://github.com/JHOFER-Cloud/helm-ci/commit/138145d6c24a5a5004bad01b83e711e85af2648f))
* add docker image ([85d41a2](https://github.com/JHOFER-Cloud/helm-ci/commit/85d41a26ed3caba5e46837c21ec0b217a9eebe1b))
* add pr_deployments ([adf725b](https://github.com/JHOFER-Cloud/helm-ci/commit/adf725b817fdc40caa851ea5e256bd6e18d9110e))
* allow custom namespace ([8ee4dfe](https://github.com/JHOFER-Cloud/helm-ci/commit/8ee4dfe0225756b4a084cd3facb087b22489e833))
* **deploy:** implement main deployment script for helm and manifest ([b5bbd75](https://github.com/JHOFER-Cloud/helm-ci/commit/b5bbd75681c0aaf516a0154b62815b10769418c7))
* **deploy:** modularize Traefik dashboard and root CA args ([c38434f](https://github.com/JHOFER-Cloud/helm-ci/commit/c38434f42db21ac21ffe4831643bc7b4eeb6d74b))
* differentiate between external and internal domain ([2adcc07](https://github.com/JHOFER-Cloud/helm-ci/commit/2adcc075f38182f838ed953089972f4da1144a8f))

# [1.0.0-dev.2](https://github.com/JHOFER-Cloud/helm-ci/compare/v1.0.0-dev.1...v1.0.0-dev.2) (2025-02-10)


### Bug Fixes

* semantic-release issue write permissions ([b288e3a](https://github.com/JHOFER-Cloud/helm-ci/commit/b288e3a4b8334523a797efbb425baf3cb646a7e1))

# 1.0.0-dev.1 (2025-02-10)


### Bug Fixes

* ca mount ([a98a7b0](https://github.com/JHOFER-Cloud/helm-ci/commit/a98a7b00d7878e7ba7241e041fff7aad7538526b))
* chart name wrong ([30ab037](https://github.com/JHOFER-Cloud/helm-ci/commit/30ab037260ac7c6a8a5831bbd5f84050f28d9578))
* **ci:** correct tag evaluation logic in Docker workflow ([17e39ba](https://github.com/JHOFER-Cloud/helm-ci/commit/17e39babf48c5fbf77637d306acaf789d29f3d23))
* custom manifest deployment ([d69e441](https://github.com/JHOFER-Cloud/helm-ci/commit/d69e4411c830e674c7ce6689d27b2494f438e962))
* deploy call ([97559af](https://github.com/JHOFER-Cloud/helm-ci/commit/97559af8ad8aea945bcc66df757983edb9ab0490))
* deployment not triggering ([1bdcaec](https://github.com/JHOFER-Cloud/helm-ci/commit/1bdcaecd8f14bb391792c492d53582f86a3e63ae))
* diff ([1712534](https://github.com/JHOFER-Cloud/helm-ci/commit/1712534e5891dca3396224c8760415c461c02703))
* forgot GitHub* in config struct ([#22](https://github.com/JHOFER-Cloud/helm-ci/issues/22)) ([669a371](https://github.com/JHOFER-Cloud/helm-ci/commit/669a371c549f5b2855c58538a9fe5e1c7a22167b))
* multiline secrets ([f6fc7bb](https://github.com/JHOFER-Cloud/helm-ci/commit/f6fc7bb0909fc417d19a24dc0d811d39728bbec7))
* pr/dev not containg cert and set values ([dcc1409](https://github.com/JHOFER-Cloud/helm-ci/commit/dcc1409f499e78da311b54c9bc2b036afa11bcb6))
* root ca volume ([89f9f35](https://github.com/JHOFER-Cloud/helm-ci/commit/89f9f35281d937ae8aa3cdaceb4f491bf312944b))
* setup root ca ([679b045](https://github.com/JHOFER-Cloud/helm-ci/commit/679b0452a87e365d041e69f3c246575fccdfb378))
* values files ([e0cd8df](https://github.com/JHOFER-Cloud/helm-ci/commit/e0cd8df553d156f1a9a6d7bc56504a961e83eb02))
* vault ingress setting not working (FIX THIS BETTER) ([e9911b6](https://github.com/JHOFER-Cloud/helm-ci/commit/e9911b6e6d6676d41227b30acf012692b48ce92b))


### Features

* add diff ([138145d](https://github.com/JHOFER-Cloud/helm-ci/commit/138145d6c24a5a5004bad01b83e711e85af2648f))
* add docker image ([85d41a2](https://github.com/JHOFER-Cloud/helm-ci/commit/85d41a26ed3caba5e46837c21ec0b217a9eebe1b))
* add pr_deployments ([adf725b](https://github.com/JHOFER-Cloud/helm-ci/commit/adf725b817fdc40caa851ea5e256bd6e18d9110e))
* allow custom namespace ([8ee4dfe](https://github.com/JHOFER-Cloud/helm-ci/commit/8ee4dfe0225756b4a084cd3facb087b22489e833))
* **deploy:** implement main deployment script for helm and manifest ([b5bbd75](https://github.com/JHOFER-Cloud/helm-ci/commit/b5bbd75681c0aaf516a0154b62815b10769418c7))
* **deploy:** modularize Traefik dashboard and root CA args ([c38434f](https://github.com/JHOFER-Cloud/helm-ci/commit/c38434f42db21ac21ffe4831643bc7b4eeb6d74b))
* differentiate between external and internal domain ([2adcc07](https://github.com/JHOFER-Cloud/helm-ci/commit/2adcc075f38182f838ed953089972f4da1144a8f))
