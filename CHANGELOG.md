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

- cluster requirements validation [#sc-4589]

### Changed

### Fixed

### Removed

### Deprecated

### Security

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
