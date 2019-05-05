# CRDB K8S Operator CHANGELOG


<a name="v0.1.2"></a>
## [v0.1.2] - 2019-05-05
### Code Refactoring
- **circleci:** streamlined build pipeline

### Bug Fixes
- circleci changed cache key to Gopkg.lock
- circleci build branches with open PRs
- circleci only build master branch
- enable circleci builds for all branches
- **circleci:** only build images for dev and master


<a name="v0.1.1"></a>
## [v0.1.1] - 2019-05-03
### Features
- added post-commit hook for changelog gen
- enable vendor cache on circleci

### Bug Fixes
- fix checkout path for circleci

### Code Refactoring
- updated version in version.go
- added some information to README.md


<a name="v0.1.0"></a>
## v0.1.0 - 2019-05-03
### Features
- added controllers for certs and crds


[Unreleased]: https://github.com/smartmachine/crdb-operator/compare/v0.1.2...HEAD
[v0.1.2]: https://github.com/smartmachine/crdb-operator/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/smartmachine/crdb-operator/compare/v0.1.0...v0.1.1
