# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

We use the following categories for changes:

- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## [Unreleased]

### Added

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.22.6] 2024-10-21

### Added

### Changed

### Fixed

- Low resource preset should take precedence over kernel 5.11 overrides [#sc-20056]

### Removed

### Deprecated

### Security

## [0.22.5] 2024-10-13

### Added

### Changed

- Update low resources presets to include new components [#sc-20036]

### Fixed

### Removed

### Deprecated

### Security

## [0.22.4] 2024-10-07

### Added

### Changed

### Fixed

- Waiting for sensor pods instead of alligator pods [#sc-19833]

### Removed

### Deprecated

### Security

## [0.22.3] 2024-09-17

### Added

- Support groundcover CLI in windows [#sc-19269]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.22.2] 2024-07-04

### Added

- support inCloud validation flow [#sc-16561]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.22.1] 2024-06-03

### Added

### Changed

- Override slice values on redeploy [#sc-16445]

### Fixed

### Removed

### Deprecated

### Security

## [0.22.0] 2024-05-22

### Added

### Changed

- Datasource key generation based on backend name instead of cluster [sc-16014]

### Fixed

### Removed

### Deprecated

### Security

## [0.21.1] 2024-05-02

### Added

- Add resources tuning for monitors manager and postgresql [sc-15557]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.21.00] 2024-03-03

### Added

### Changed

### Fixed

### Removed

### Deprecated

### Security

- apply security patches and use go1.22 [#sc-14028]

## [0.20.21] 2024-02-15

### Added

- add `version` flag to `deploy` command  [#sc-13642]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.20] 2024-01-11

### Added

### Changed

### Fixed

- CLI checks expected PVC amount dynamically [#sc-12837]

### Removed

### Deprecated

### Security

## [0.10.19] 2023-11-21

### Added

### Changed

### Fixed

- fixed get datasources API key command issues and added cluster picker [#sc-11777]

### Removed

### Deprecated

### Security

## [0.10.18] 2023-11-15

### Added

- added get datasources API key command [#sc-11671]
- added get client token command [#sc-11670]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.17] 2023-11-12

### Added

### Changed

### Fixed

- fixed bug with large and huge presets [#sc-11599]

### Removed

### Deprecated

### Security

## [0.10.16] 2023-11-12

### Added

### Changed

- increased otel-collector memory and cpu limits [#sc-10997]
- increased k8swatcher memory and cpu limits [#sc-11513]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.15] 2023-10-08

### Added

### Changed

- revert kube-state-metrics and custom metrics scraping by default [#sc-10892]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.14] 2023-10-03

### Added

### Changed

- changed low-resources values [#sc-10885]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.13] 2023-09-20

### Added

### Changed

- support kube-state-metrics and custom metrics scraping by default [#sc-10432]
- generate grafana service account token command [#sc-10714]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.12] 2023-08-23

### Added

- add `storage-class` flag to `deploy` command  [#sc-10362]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.11] 2023-08-01

### Added

### Changed

### Fixed

- support storageclass beta annotation [#sc-10131]

### Removed

### Deprecated

### Security

## [0.10.10] 2023-08-01

### Added

- support quay registry [#sc-10019]
- add cluster storage provision validation [#sc-9988]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.9] 2023-07-24

### Added

### Changed

### Fixed

- segment fixes [#sc-9870]

### Removed

### Deprecated

### Security

## [0.10.8] 2023-07-18

### Added

### Changed

### Fixed

- segment fixes [#sc-9870]

### Removed

### Deprecated

### Security

## [0.10.7] 2023-07-16

### Added

### Changed

### Fixed

### Removed

- low resources mode message [#sc-9762]

### Deprecated

### Security

## [0.10.6] 2023-07-13

### Added

- arm support [#sc-9745]

### Changed

### Fixed

- fix tenants prompt [#sc-9748]

### Removed

### Deprecated

### Security

## [0.10.5] 2023-07-11

### Added

- allow override tolerations on agent via values [#sc-9689]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.4] 2023-07-10

### Added

### Changed

- tune vmagent resources [#sc-9511]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.3] 2023-07-04

### Added

- added a "huge" resources presets for over 100 nodes clusters, removed medium [#sc-9569]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.10.2] 2023-07-04

### Added

### Changed

- increase clickhouse memory limits [#sc-9558]
- increase opentelemetry-collector low resources memory preset [#sc-8916]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.1] 2023-06-27

### Added

### Changed

- tune agent medium preset [#sc-9454]

### Fixed

### Removed

### Deprecated

### Security

## [0.10.0] 2023-06-27

### Added

- support multi tenancy deployment [#sc-9443]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.9.9] 2023-05-30

### Added

- validation support standalone agent and backend installation [#sc-8224]

### Changed

### Fixed

### Removed

### Deprecated

### Security

- upgrade go modules [#sc-8190]

## [0.9.8] 2023-05-25

### Added

### Changed

### Fixed

- otel low resources calibration [#sc-8658]

### Removed

### Deprecated

### Security

## [0.9.7] 2023-05-14

### Added

### Changed

- increase clickhouse mem limit to 2g [#sc-8481]

### Fixed

- print low-resources mode message on backend [#sc-8481]

### Removed

### Deprecated

### Security

## [0.9.6] 2023-05-08

### Added

### Changed

### Fixed

- clickhouse emptydir mode [#sc-8445]

### Removed

### Deprecated

### Security

## [0.9.5] 2023-05-08

### Added

### Changed

- increase low preset of clickhouse memory [#sc-8405]

### Fixed

### Removed

### Deprecated

### Security

## [0.9.4] 2023-05-08

### Added

### Changed

### Fixed

- fixed pvc creation check bug [#sc-8388]

### Removed

### Deprecated

### Security

## [0.9.3] 2023-05-07

### Added

- tune clickhouse and otel-collector resources [#sc-8227]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.9.2] 2023-04-24

### Added

- Added high cluster preset [#sc-8158]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.9.1] 2023-04-18

### Added

### Changed

### Fixed

- ignore incompatible nodes in kernel versions [#sc-8109]

### Removed

### Deprecated

### Security

## [0.9.0] 2023-04-17

### Added

- add `mode` flag to `deploy` command [#sc-8011]
- legacy mode on outdated kernel version [#sc-8011]
- kube state metrics flag [#sc-8036]
- kernel 5.11 agent memory limits [#sc-8063]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.8.7] 2023-03-27

### Added

- custom metrics flag [#sc-7824]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.8.6] 2023-03-23

### Added

### Changed

- removed promscale resources, and updated shepherd-collector's [#sc-7727]

### Fixed

### Removed

### Deprecated

### Security

## [0.8.5] 2023-03-13

### Added

### Changed

- updated wait time for alligators to run [#sc-7650]

### Fixed

### Removed

### Deprecated

### Security

## [0.8.4] 2023-03-13

### Added

### Changed

- updated wait time for alligators to run [#sc-7651]

### Fixed

### Removed

### Deprecated

### Security

## [0.8.3] 2023-03-06

### Added

### Changed

- use new cluster list schema [#sc-7574]

### Fixed

### Removed

### Deprecated

### Security

## [0.8.2] 2023-02-27

### Added

- added tags to report [#sc-6938]
- partial success flow [#sc-7483]

### Changed

### Fixed

- suppress analytics logger [#sc-7482]

### Removed

### Deprecated

### Security

## [0.8.1] 2023-02-21

### Added

### Changed

### Fixed

- skip iat check [#sc-7376]

### Removed

### Deprecated

### Security

## [0.8.0] 2023-02-21

### Added

- add analytics [#sc-7298]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.7.1] 2023-02-17

### Added

### Changed

### Fixed

- fix fetch api-key [#sc-7357]

### Removed

### Deprecated

### Security

## [0.7.0] 2023-02-16

### Added

- add `token` flag to `deploy` command [#sc-7216]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.6.0] 2023-02-05

### Added

- add support to arm64 agent [#sc-7116]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.5.4] 2023-01-31

### Added

### Changed

- renamed uninstall command to delete [#sc-7033]

### Fixed

### Removed

### Deprecated

### Security

## [0.5.1] 2023-01-16

### Added

### Changed

### Fixed

- no slack link on fail [#sc-6903]

### Removed

### Deprecated

### Security

## [0.5.0] 2023-01-16

### Added

- Added --no-pvc option [#sc-6887]

### Changed

### Fixed

- support aws eks 1.23+ with no ebs driver installed

### Removed

### Deprecated

### Security

## [0.4.3] 2023-01-08

### Added

### Changed

- Improved logging [#sc-6298]

### Fixed

### Removed

### Deprecated

### Security

## [0.4.2] 2023-01-08

### Added

- Add store-all-logs flag [#sc-6717]

### Changed

### Fixed

### Removed

### Deprecated

### Security

## [0.4.1] 2023-01-03

### Added

- remove cpu and memory requirements [#sc-6682]

### Fixed

- resources limit threshold optimaztions [#sc-6682]
- deep merge default chart values [#sc-6686]

## [0.4.0] 2023-01-02

### Added

- add low-resources preset [#sc-5895]
- add low-resources flag to `deploy` [#sc-5895]
- support minikube and kind clusters [#sc-5895]
- add shepherd resources tune [#sc-6617]

### Changed

- added a SaaS link on helm validation error [#sc-6297]

## [0.3.18] 2022-12-18

### Added

- print re-run message after cli upgrade [#sc-6361]

### Fixed

- values override deep copy [#sc-6358]

## [0.3.17] 2022-12-07

### Added

- helm reuse only user supplied values [#sc-6189]
- increase pvc bound timeout [#sc-6188]

## [0.3.16] 2022-12-01

### Added

- add retry mechanism [#sc-5170]

### Fixed

- use slice contains in GetTolerableNodes [#sc-6056]

## [0.3.15] 2022-11-11

### Fixed

- merge values override recursive [#sc-5644]

## [0.3.14] 2022-10-20

### Added

- add experimental flag to `deploy` [#sc-5250]

## [0.3.13] 2022-10-13

### Added

- prompt taints for tolerations approval [#sc-5028]

## [0.3.12] 2022-10-12

### Added

- improved uninstall command [#sc-4839]

### Fixed

- klog use dedicated flagset [#sc-5172]

## [0.3.11] 2022-10-06

### Added

- command time took metric [#sc-4963]
- add spinner to cli update [#sc-5038]
- add aws cli version validation [#sc-4942]
- spinner handle interrupt signal [#sc-5092]

## [0.3.10] 2022-09-21

### Added

- added watch pvc readiness [#sc-4860]
- validate node schedulable [#sc-4878]

### Changed

- change timeout for portal connection polling to 7 mins [#sc-4860]

## [0.3.9] 2022-09-20

### Added

- override departed authentication api version [#sc-4856]
- cluster established connectivity validation [#sc-4861]

## [0.3.8] 2022-09-19

### Fixed

- handle private mail error [#sc-4852]

## [0.3.7] 2022-09-19

### Added

- authentication validation [#sc-4815]

## [0.3.4] 2022-09-14

### Added

- store helm storage in groundcover cache directory [#sc-4780]
- disable klog stderr output [#sc-4775]

## [0.3.3] 2022-09-13

### Added

### Changed

### Fixed

- added k8s client auth plugins

### Removed

### Deprecated

### Security

## [0.3.2] 2022-09-12

### Added

- interrupt signal handler [#sc-4649]
- print validation errors in ui [#sc-4660]

## [0.3.1] 2022-09-07

### Fixed

- cluster report metrics fix [#sc-4657]

## [0.3.0] 2022-09-07

### Added

- cluster requirements validation [#sc-4589]

## [0.2.10] 2022-08-31

### Changed

- api ListCluster endpoint new schema [#sc-4510]

## [0.2.9] 2022-08-30

### Added

- add values override flag to `deploy` command [#sc-4395]
- support reuse values on upgrade [#sc-4420]
- add git integration flags to `deploy` command [#sc-4412]

## [0.2.8] 2022-08-25

### Added

- set resources based on allocation [#sc-4312]

### Fixed

- set version on cmd.BinaryVersion [#sc-4400]

## [0.2.6] 2022-08-22

### Changed

- split authentication functionality to api and auth clients [#sc-4313]

### Fixed

- validate or refresh auth0 token per api request [#sc-4313]

## [0.2.5] 2022-08-17

### Added

- capture self-update metrics [#sc-4261]

### Fixed

- add help cmd to skip auth [#sc-4255]
- ignore unknown command error in metrics [#sc-4256]
- Auth0 device code confirmation polling interval and timeout [#sc-4170]

## [0.2.4] 2022-08-16

### Added

- `auth print-api-key` command to print groundcover api-key [#sc-4121]
- `auth login` command to authenticate your accout in groundcover [#sc-4121]

## [0.2.3] 2022-08-07

### Changed

- Improve metrics collection [#sc-3674]
- More tolerant deployment timeout [#sc-3674]
- kubeconfig missing error more informative [#sc-4038]

## [0.2.2] 2022-08-02

### Added

- Node minimum requirements mechanism [#sc-3920]
- Support KUBECONFIG environment variable override [#sc-3826]

## [0.2.1] 2022-07-20

### Security

- Upgrade dependencies to resolve security vulnerabilities [#sc-3755]

## [0.2.0] 2022-07-20

### Added

- Add assume-yes flag [#sc-3193]
- Add kube-context flag [#sc-3397]

### Changed

- Use Helm V3 SDK [#sc-3397]
- Use Kubernetes SDK [#sc-3242]

## [0.1.0] - 2022-07-04

### Added

- `login` command to authenticate your accout in groundcover
- `deploy` command to install groundcover helm chart
- `uninstall` command to remove groundcover helm release
- `status` command to get current deployment status
