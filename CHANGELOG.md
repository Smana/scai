# Changelog

## [0.5.2](https://github.com/Smana/scai/compare/v0.5.1...v0.5.2) (2025-10-19)


### Documentation

* fix installation instructions and replace example URLs ([81e9305](https://github.com/Smana/scai/commit/81e9305fa9fc82aead9075e33f154956600ab393))
* fix installation instructions for tar.gz archives ([c05fe5d](https://github.com/Smana/scai/commit/c05fe5d01a9d0ce0708bfa412b9eefad9e14801c))
* replace Arvo-AI repository references with generic placeholder ([9bcb6c6](https://github.com/Smana/scai/commit/9bcb6c674eb0cfb247bb3265a0de022f065db87f))

## [0.5.1](https://github.com/Smana/scia/compare/v0.5.0...v0.5.1) (2025-10-18)


### Bug Fixes

* restore essential OpenTofu modules and fix YAML parsing ([bbf054a](https://github.com/Smana/scia/commit/bbf054a512d8ced1617f937950f4f30d8fc746a1))
* restore essential OpenTofu modules and fix YAML parsing ([f36182b](https://github.com/Smana/scia/commit/f36182b5e5225b00c27a42f1d8e862c1fb2f6e1a))

## [0.5.0](https://github.com/Smana/scia/compare/v0.4.0...v0.5.0) (2025-10-18)


### Features

* implement 3-tier rule-based decision engine for deployment strategy ([8fed5e0](https://github.com/Smana/scia/commit/8fed5e021ea950026d798a806ded96355d9971c8))

## [0.4.0](https://github.com/Smana/scia/compare/v0.3.0...v0.4.0) (2025-10-18)


### Features

* add health checks, LLM tracking, and deployment improvements ([50a1841](https://github.com/Smana/scia/commit/50a1841498201d6de9156f3efff479fe271ee039))
* add health checks, LLM tracking, and deployment improvements ([6b7e517](https://github.com/Smana/scia/commit/6b7e5176c7f027901a92075c6a95313a185b516b))
* add OpenTofu modules for CI validation ([c31e0dc](https://github.com/Smana/scia/commit/c31e0dc1e7fbd7b50af7b354b110a2863f7ae382))


### Bug Fixes

* streamline pre-commit config for Dagger CI compatibility ([b8c1ac5](https://github.com/Smana/scia/commit/b8c1ac5f615626f65333235d5e3d9df42a632062))
* update AWS provider to ~&gt; 6.0 for EKS and Lambda compatibility ([22ae6a0](https://github.com/Smana/scia/commit/22ae6a008c15fbe3714cc94491ce25817f14dd12))
* update Terraform module versions for AWS provider 6.x compatibility ([1fe8d62](https://github.com/Smana/scia/commit/1fe8d62e38ae995275b53a6d2bf85cc9922f2e3a))
* update Terraform module versions for AWS provider 6.x compatibility ([d31c063](https://github.com/Smana/scia/commit/d31c0639f92b79dd67a7bfd31f51636ecb830153))


### Code Refactoring

* reduce cyclomatic complexity in SQLiteStore.List ([d25be2d](https://github.com/Smana/scia/commit/d25be2d76b3cb86ef1bcabb0aaa9068fe77862e7))


### Continuous Integration

* add OpenTofu validation using Dagger ([032a290](https://github.com/Smana/scia/commit/032a2909ca12b862c125fa227833ad7a54b71c72))

## [0.3.0](https://github.com/Smana/scia/compare/v0.2.0...v0.3.0) (2025-10-18)


### Features

* add deployment tracking, UI consistency, and LLM improvements ([6d7065d](https://github.com/Smana/scia/commit/6d7065d8192a85a2dff478cabc4f1a6b570ee279))
* add deployment tracking, UI consistency, and LLM improvements ([89471b8](https://github.com/Smana/scia/commit/89471b8ae75c83aad8563eb4651319d3902dd863))

## [0.2.0](https://github.com/Smana/scia/compare/v0.1.0...v0.2.0) (2025-10-17)


### Features

* add Ollama LLM integration with Docker support and natural language parsing ([617e29c](https://github.com/Smana/scia/commit/617e29ce951345417bd56ff47ee809bac9ab6e4c))
* Add Ollama LLM integration with Docker support and natural language parsing ([a81ec4a](https://github.com/Smana/scia/commit/a81ec4a0e53d6f21e8a9cacfebbbd05e4cbeaa27))
* **ci:** first version of CI using dagger taskfile goreleaser ([aa44acd](https://github.com/Smana/scia/commit/aa44acd954f9f56c9c2f3bfb62c754f857e81703))
* initial commit ([5839026](https://github.com/Smana/scia/commit/583902656162630a774468398e7e45b0711be2a9))
* upgrade golang version to v1.25 ([33160ed](https://github.com/Smana/scia/commit/33160ed7f80791552a67abb9506a726263e77ff4))


### Bug Fixes

* add framework handler constants to resolve goconst warnings ([236b7a1](https://github.com/Smana/scia/commit/236b7a1db57881d4da82f0402c26b22bb942674a))
* **ci:** simplify commitlint config to resolve validation errors ([093770a](https://github.com/Smana/scia/commit/093770a6c570e7cecac65cff6f59715ff66d3deb))
* correct goimports formatting for import groups ([548821d](https://github.com/Smana/scia/commit/548821d4feb8427663933216b41a1c83368cd600))
* improve LLM instance type detection to preserve exact specifications ([e235031](https://github.com/Smana/scia/commit/e23503175ca678737a884fbae7b752d879437793))
* resolve errcheck and goconst lint errors ([6d1f626](https://github.com/Smana/scia/commit/6d1f62619bc51a8c61db5df8d897d2c2ca4fbeb7))
* resolve lint errors (goconst, gocyclo, gosec, octal literals) ([8c16e8d](https://github.com/Smana/scia/commit/8c16e8dd3e1584aa2904262ea4b9acffe2d8afd9))
* resolve remaining lint errors (goconst, gosec, formatting) ([474d033](https://github.com/Smana/scia/commit/474d0336af0e6c43ea11704d882f92f1c4e5e0d3))


### Documentation

* add security policy and handle Ollama CVEs ([599096b](https://github.com/Smana/scia/commit/599096b194abb29494d527b64776bbae8d6eead9))
* drastically simplify README for better usability ([25d9556](https://github.com/Smana/scia/commit/25d9556dce86572ce6afe9beaea85ee213d7fd64))


### Code Refactoring

* apply security, code quality, and linting improvements ([c49d499](https://github.com/Smana/scia/commit/c49d499da91550834821960e8cd8145e6872109a))
