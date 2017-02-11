# Mdindicator

`mdindicator` provides a facility to watch a markdown file, and serve it as an HTML.

## Install

To install, use `go get`:

```bash
$ go get -u github.com/elpinal/mdindicator
```

-u flag stands for "update".

## Examples

Basic usage:

```bash
$ mdindicator README.md
```

then, see http://localhost:8080.

To change port:

```bash
$ mdindicator -http :8888 README.md
```

## Contribution

1. Fork ([https://github.com/elpinal/mdindicator/fork](https://github.com/elpinal/mdindicator/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[elpinal](https://github.com/elpinal)
