# ged

genealogical data tooling

This tool lets you:
- calculate relationships between family members

## Development

Some auxilliary tools required:
- [golang](https://go.dev/), for building the application from source
- [just](https://just.systems), to streamline code building and other tasks

The justfile defines the most common operations.

```sh
# list tasks
$ just

# get application dependencies
$ just deps

# build binary
$ just build
```

### Operational tasks

#### Tests

```sh
$ just test

# run tests on a specific package path
$ just test ./internal/repo/...

# run tests with some flags
$ just ARGS='-v -count=1 -failfast' test

# run tests on a specific package path, while also specifying flags
$ just ARGS='-v -run TestFoo' test ./internal/gedcom7/...
```

#### Static analysis

```sh
$ just vet
```
