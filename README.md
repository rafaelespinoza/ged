# ged

[![build](https://github.com/rafaelespinoza/ged/actions/workflows/build.yaml/badge.svg)](https://github.com/rafaelespinoza/ged/actions/workflows/build.yaml)

genealogical data tooling

## Features

- Ingest [GEDCOM](https://gedcom.io) data, a defacto standard file format spec for genealogical data
- Calculate relationships between family members via common ancestors
- View, peruse records in the terminal
- Draw a family tree from GEDCOM data
- Structured, leveled logging messages written to standard error

## Getting started

Required tooling:
- [golang](https://go.dev), for building the application from source
- [just](https://just.systems), to streamline code building and other tasks

Optional, but recommended tooling:
- [bkt](https://github.com/dimo414/bkt), to cache views when interactively exploring data
- [fzf](https://github.com/junegunn/fzf), to enhance calculating relationships between people
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

Take GEDCOM data as input and output a family tree as an SVG or PNG.
```sh
# By default, it renders a SVG
$ bin/main draw < testdata/simpsons.ged > /tmp/ged/simpsons.svg

# Render a PNG
$ bin/main draw -output-format=png < testdata/simpsons.ged > /tmp/ged/simpsons.png
```
Under the hood, a [Mermaid flowchart](https://mermaid.js.org/syntax/flowchart.html) is constructed from the GEDCOM data. Then that flowchart is rendered into a standard image format.

Another output format is the Mermaid flowchart itself. The use case here is for
any manual edits you may want to do before rendering it again.
```sh
$ bin/main draw -output-format=mermaid < testdata/simpsons.ged > /tmp/ged/simpsons.mermaid

# ... manual adjustments to Mermaid flowchart file ...

# to render the manually-modified Mermaid flowchart, just specify the input-format
$ bin/main draw -input-format=mermaid < /tmp/ged/simpsons.mermaid > /tmp/ged/simpsons.svg
```

### explore-data 

#### relate

Calculate relationship between 2 people.

If you have [fzf](https://github.com/junegunn/fzf) on your system, then use fuzzy search to select the people.
```sh
$ ./explore-data.sh relate testdata/kennedy.ged
```

If you don't have fzf or don't want to use it, specify the IDs of the people like so:
```sh
$ bin/main explore-data relate -p1 @I0@ -p2 @I10@ < testdata/kennedy.ged
```

#### show

Display details on a person in a simple group sheet view.

If you have [fzf](https://github.com/junegunn/fzf) on your system, then use fuzzy search to select the people.
```sh
$ ./explore-data.sh show testdata/kennedy.ged
```

If you don't have fzf or don't want to use it, specify the ID of the person like so:
```sh
$ bin/main explore-data show -target-id @I10@ < testdata/kennedy.ged
```
