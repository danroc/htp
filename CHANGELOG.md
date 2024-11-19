# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.4] - 2023-01-31

## [0.1.3] - 2022-11-07

### Fixed

- Make sure that HTTP connection can be reused

## [0.1.2] - 2022-11-06

## [0.1.1] - 2022-05-01

## [0.1.0] - 2022-04-29

### Added

- Add 'adjust mode' to output offset
- Add a default timeout
- Add default protocol (`https`)
- Add option to format output
- Add option to sync system time
- Add quiet option
- Add sleep time
- Output raw offset to support other platforms
- Show request number in trace

### Changed

- Add 'Cache-Control: no-cache' header
- Change available command line options
- Change default format to ISO 8601
- Change default number of requests
- Change default time format layout
- Change default URL to `https://www.google.com`
- Display offset at each loop
- Don't follow redirects
- Fail if server time changes
- Improve arguments parsing
- Print corrected date instead of offset
- Print current time

### Fixed

- Update sync.bat permissions
- Hide usage when an error occurs

[Unreleased]: https://github.com/danroc/htp/compare/v0.1.4...HEAD
[0.1.4]: https://github.com/danroc/htp/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/danroc/htp/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/danroc/htp/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/danroc/htp/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/danroc/htp/releases/tag/v0.1.0
