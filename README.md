# ged

genealogical data tooling

## Features

- Ingest [GEDCOM](https://gedcom.io) data, a defacto standard file format spec for genealogical data
- Calculate relationships between family members via common ancestors
- Draw a family tree as a [Mermaid flowchart](https://mermaid.js.org/syntax/flowchart.html)
- Structured, leveled logging messages written to standard error

## Getting started

Required tooling:
- [golang](https://go.dev), for building the application from source
- [just](https://just.systems), to streamline code building and other tasks

Optional, but recommended tooling:
- [fzf](https://github.com/junegunn/fzf), to enhance calculating relationship between people
- [jq](https://jqlang.github.io/jq/manual), view parsed data as JSON

The justfile defines the most common operations.

```sh
# list tasks
$ just

# get, tidy, vendor application dependencies
$ just mod

# build, run binary
$ just build
$ just run
$ just buildrun
```

### Operational tasks

Tests

```sh
$ just test

# run tests on a specific package path
$ just test ./internal/repo/...

# run tests with some flags
$ just ARGS='-v -count=1 -failfast' test

# run tests on a specific package path, while also specifying flags
$ just ARGS='-v -run TestFoo' test ./internal/gedcom/...
```

Static analysis

```sh
$ just vet
```

## Usage

Presumably, you have some of your own GEDCOM data. If you don't, there are
samples in `testdata/`. These examples use that testdata.

Optionally, establish a working directory.
```sh
$ mkdir /tmp/ged && chmod -v 700 /tmp/ged
```

### draw

Generate a family tree as a Mermaid flowchart.
```sh
$ bin/main draw < testdata/simpsons.ged | tee /tmp/ged/simpsons.mermaid
```
The resulting file can be input to a Mermaid interpreter, such as https://mermaid.live.

### relate

Calculate relationship between 2 people.

If you have [fzf](https://github.com/junegunn/fzf) on your system, then use fuzzy search to select the people.
```sh
$ bin/main relate < testdata/kennedy.ged
```

If you don't have fzf or don't want to use it, specify the IDs of the people like so:
```sh
$ bin/main relate -p1 @I111@ -p2 @I222@ < testdata/kennedy.ged
```
