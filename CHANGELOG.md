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

- added tags to report [#sc-6938]
- partial success flow [#sc-7483]

### Changed

### Fixed

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
