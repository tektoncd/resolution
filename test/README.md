# End-to-End Tests

## Running All E2E Tests

To run e2e tests call the following `go` command:

```bash
$ go test -tags=e2e ./test/...
```

## Running a Single E2E Test

Each test is given its own directory to avoid all the test code sharing
a single module namespace and accidentally clobbering each others
helper functions and types.

To run a single test on its own, use the path to the its directory. For
example, run just `smoke_test`:

```bash
$ go test -tags=e2e ./test/smoke_test/...
```
