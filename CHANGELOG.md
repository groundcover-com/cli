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

- `auth print-api-key` command to print groundcover api-key [#sc-4121]
- `auth login` command to authenticate your accout in groundcover [#sc-4121]

### Changed

### Fixed

- Auth0 device code confirmation polling interval and timeout [#sc-4170]

### Removed

### Deprecated

### Security

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
