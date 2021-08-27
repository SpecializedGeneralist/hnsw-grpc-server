# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
- Update `hnswlib` code to release [v0.5.2](https://github.com/nmslib/hnswlib/releases/tag/v0.5.2).
- Use Go version `1.17`.
- Upgrade dependencies.

## [1.0.0] - 2021-05-01
### Added
- Significantly increased data durability with the introduction of
  Write-Ahead Logging (WAL).

  A record of all operations performed on an index is appended to a Write-Ahead
  Log (WAL) file (see `pkg/wal` package). The log is emptied only after each
  successful saving operation, that is, once the in-memory content of the index
  is "securely" persisted on disk.
  If a log file is present, any recorded operation is re-applied while loading
  an existing index.

### Changed
- **Breaking**: HNSW internal state structure changed in a back-incompatible way.

  The internal state (i.e. `type hnswState`), which is persisted to disk when
  saving an index, now contains all the index configuration parameters.
  This might be convenient for indices inspection and recovery.
- Refactor and improve test cases.

### Removed
- File `hnsw_wrapper.o`, which is part of the building process.

### Fixed
- Fix bug causing re-loaded indices to break when inserting new vectors.

## [0.2.0] - 2021-04-28
### Added
- Tests.
- GitHub Workflow CI.
- This CHANGELOG file.

### Changed
- Improved indices saving and loading implementation.
- Improved logging.
- Significant code refactoring.

## [0.1.1] - 2021-04-13
### Added
- Methods `InsertVectorWithId` and `InsertVectorsWithIds`.

## [0.1.0] - 2021-04-05
### Added
- First release.
