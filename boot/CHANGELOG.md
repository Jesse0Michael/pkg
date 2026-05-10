# Changelog

All notable changes to `github.com/jesse0michael/pkg/boot` will be documented in this file by Release Please.

## [1.2.0](https://github.com/Jesse0Michael/pkg/compare/boot/v1.1.0...boot/v1.2.0) (2026-05-10)


### Features

* support cli args in config ([46606a3](https://github.com/Jesse0Michael/pkg/commit/46606a33fe6442d2a5e066e186cbac7adc6b3f37))

## [1.1.0](https://github.com/Jesse0Michael/pkg/compare/boot/v1.0.0...boot/v1.1.0) (2026-05-09)


### Features

* process config  by attributes, use boot options ([420ba96](https://github.com/Jesse0Michael/pkg/commit/420ba965c2357bad66f8063749f3f04a2c5691a3))

## [1.0.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.7.0...boot/v1.0.0) (2026-05-09)


### ⚠ BREAKING CHANGES

* NewLogger now accepts a logger.Config struct instead of reading environment variables directly. LogLevel, LogOutput, LogSource, and LogFormat are now methods on Config.

### Features

* require Config parameter for logger.NewLogger ([0bbb1ff](https://github.com/Jesse0Michael/pkg/commit/0bbb1ffcbe30bc92d8780fed2c4468f0e557934e))

## [0.7.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.6.0...boot/v0.7.0) (2026-04-25)


### Features

* add data package ([0d25b8a](https://github.com/Jesse0Michael/pkg/commit/0d25b8af0ecc3403ad3d50f132ba1360b2acce0c))


### Bug Fixes

* go version 1.26.2 ([0416552](https://github.com/Jesse0Michael/pkg/commit/0416552a892ba310ac2ff1b3809f40aabd4caa3b))

## [0.6.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.5.1...boot/v0.6.0) (2026-02-24)


### Features

* setup open telemetry ([3b2fee8](https://github.com/Jesse0Michael/pkg/commit/3b2fee8401fe6379f560999724910589fe7cadc9))

## [0.5.1](https://github.com/Jesse0Michael/pkg/compare/boot/v0.5.0...boot/v0.5.1) (2026-02-24)


### Bug Fixes

* reorder app Run ([b344b63](https://github.com/Jesse0Michael/pkg/commit/b344b63976927872f701d89185d73993f8dce9af))

## [0.5.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.4.0...boot/v0.5.0) (2026-02-23)


### Features

* add OtelLogProvider ([ade8995](https://github.com/Jesse0Michael/pkg/commit/ade8995fff10b2010e8f29b3169e3f12086042a3))
* shutdown providers ([7592be3](https://github.com/Jesse0Michael/pkg/commit/7592be39bb9bf3aac99336938882a6f6f3c64b0d))

## [0.4.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.3.0...boot/v0.4.0) (2026-02-18)


### Features

* rate limiter interceptor ([594def4](https://github.com/Jesse0Michael/pkg/commit/594def4a92083b95e363cc880cfc5bfb195e4855))

## [0.3.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.2.0...boot/v0.3.0) (2026-02-16)


### Features

* add boot package ([0692e90](https://github.com/Jesse0Michael/pkg/commit/0692e908c569716b1c4fb378a68ae3b98698b6c6))
* boot app ([a870daa](https://github.com/Jesse0Michael/pkg/commit/a870daa55773e5b0d6d99b7ff717805fda1925cf))
* boot app and run ([21553ac](https://github.com/Jesse0Michael/pkg/commit/21553ac3a4b9877ef7068da0c78a4426a410fd4f))
* sync changelogs ([5252169](https://github.com/Jesse0Michael/pkg/commit/52521696340ea3310e6dd49726fbb5207d8d5cc1))


### Bug Fixes

* force tag ([b4970ec](https://github.com/Jesse0Michael/pkg/commit/b4970ec82f9d801e4345d61b5515d4c2b6c0c6e2))
* make tidy ([94dc3b9](https://github.com/Jesse0Michael/pkg/commit/94dc3b94503d4ab715c363008eba9e79036830c6))

## [0.2.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.1.1...boot/v0.2.0) (2026-02-16)


### Features

* add boot package ([0692e90](https://github.com/Jesse0Michael/pkg/commit/0692e908c569716b1c4fb378a68ae3b98698b6c6))
* boot app ([a870daa](https://github.com/Jesse0Michael/pkg/commit/a870daa55773e5b0d6d99b7ff717805fda1925cf))
* boot app and run ([21553ac](https://github.com/Jesse0Michael/pkg/commit/21553ac3a4b9877ef7068da0c78a4426a410fd4f))
* sync changelogs ([5252169](https://github.com/Jesse0Michael/pkg/commit/52521696340ea3310e6dd49726fbb5207d8d5cc1))


### Bug Fixes

* force tag ([b4970ec](https://github.com/Jesse0Michael/pkg/commit/b4970ec82f9d801e4345d61b5515d4c2b6c0c6e2))
* make tidy ([94dc3b9](https://github.com/Jesse0Michael/pkg/commit/94dc3b94503d4ab715c363008eba9e79036830c6))

## [0.1.1](https://github.com/Jesse0Michael/pkg/compare/boot/v0.1.0...boot/v0.1.1) (2026-02-15)


### Bug Fixes

* make tidy ([94dc3b9](https://github.com/Jesse0Michael/pkg/commit/94dc3b94503d4ab715c363008eba9e79036830c6))

## [0.1.0](https://github.com/Jesse0Michael/pkg/compare/boot/v0.0.0...boot/v0.1.0) (2025-11-15)


### Features

* sync changelogs ([5252169](https://github.com/Jesse0Michael/pkg/commit/52521696340ea3310e6dd49726fbb5207d8d5cc1))


### Bug Fixes

* force tag ([b4970ec](https://github.com/Jesse0Michael/pkg/commit/b4970ec82f9d801e4345d61b5515d4c2b6c0c6e2))

## 0.0.0

### Features

- add the boot package foundation ([0692e90](https://github.com/Jesse0Michael/pkg/commit/0692e908c569716b1c4fb378a68ae3b98698b6c6))
- add initial boot application wiring ([a870daa](https://github.com/Jesse0Michael/pkg/commit/a870daa55773e5b0d6d99b7ff717805fda1925cf))
- add run support to the boot app ([21553ac](https://github.com/Jesse0Michael/pkg/commit/21553ac3a4b9877ef7068da0c78a4426a410fd4f))
