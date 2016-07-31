# graceful-restart
[![Build Status](https://travis-ci.org/rogerclotet/graceful-restart.svg?branch=master)](https://travis-ci.org/rogerclotet/graceful-restart)

This is an effort to show a way of doing live code deployment with 0 downtime in a CQRS environment.

The example is very simple, and the saving and restoring the snapshot is so fast you wouldn't notice any change if it wasn't handled well. To better test it you can add a `time.Sleep(10*time.Second)` and see the commands returning `200 OK` but not changing anything until 10s later, and queries waiting for the snapshot to load.

## Requirements
- go 1.7

## Notes
- SIGHUP causes the program to save a snapshot, execute a child process which starts receiving requests, and then restores the snapshot.
- While the snapshot is being loaded the commands and queries are enqueued, and as soon as it's done they start being handled.

## TO DO
- Tests, benchmarks, and more tests.
- Clarify the `main.go` file and maybe move some of the logic to its own package.