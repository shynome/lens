# Changelog

## [1.3.4] - 2022-10-12

### Fix

- sent a comment when sse connected for pass proxy

## [1.3.3] - 2022-10-12

### Fix

- add the missing header of sse api

## [1.3.2] - 2022-10-11

### Fix

- Call api handler register event listener is not clean when the event is not deal, now always close event listener after req context done

## [1.3.1] - 2022-10-11

### Fix

- sse subscribe failed with timeout middleware

## [1.3.0] - 2022-10-09

### Change

- rewrite with echo

## [1.2.1] - 2022-09-25

### Added

- fix cors

## [1.2.0] - 2022-09-25

### Added

- add cors headers

## [1.1.0] - 2022-09-22

### Added

- add query field `w`(WorkerToken) to limit others impersonating your normal worker
