# CRDB K8S Operator CHANGELOG

<a name="unreleased"></a>
## [Unreleased]
### Bug Fixes
- circleci only build master branch
- enable circleci builds for all branches


<a name="v0.1.1"></a>
## [v0.1.1] - 2019-05-03
### Bug Fixes
- fix checkout path for circleci

### Code Refactoring
- updated version in version.go
- added some information to README.md

### Features
- added post-commit hook for changelog gen
- enable vendor cache on circleci


<a name="v0.1.0"></a>
## v0.1.0 - 2019-05-03
### Features
- added controllers for certs and crds


[Unreleased]: https://github.com/smartmachine/crdb-operator/compare/v0.1.1...HEAD
[v0.1.1]: https://github.com/smartmachine/crdb-operator/compare/v0.1.0...v0.1.1
