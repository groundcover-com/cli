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
