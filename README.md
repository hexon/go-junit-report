# go-junit-report

Converts `go test` output to an xml report, suitable for applications that
expect junit xml reports (e.g. [Jenkins](http://jenkins-ci.org)).

[![Build Status][travis-badge]][travis-link]
[![Report Card][report-badge]][report-link]

## Installation

Go version 1.11 or higher (with module support) is required. Install using the
`go get` command:
```bash
go get github.com/hexon/go-junit-report
```

## Usage

go-junit-report reads the `go test` verbose output from standard in and writes
junit compatible XML to standard out.

```bash
go test -v ./... 2>&1 | go-junit-report -set-exit-code > report.xml
```

Note that it also can parse benchmark output with `-bench` flag:
```bash
go test -v -bench . ./... 2>&1 | go-junit-report > report.xml
```

Command line flags:
```
Usage of go-junit-report:
  -full-package-classname
        use the full package name as the test classname instead of just the last part
  -go-version string
        specify the value to use for the go.version property in the generated XML
  -no-xml-header
        do not print xml header
  -package-name string
        specify a package name (compiled test have no package name in output)
  -set-exit-code
        set exit code to 1 if tests failed
  -strip-ansi-escape-codes
        strip ANSI escape codes (terminal color codes)
```

## Contribution

Create an Issue and discuss the fix or feature, then fork the package.
Fix or implement feature. Test and then commit change.
Specify #Issue and describe change in the commit message.
Create Pull Request. It can be merged by owner or administrator then.

This is a fork of [github.com/jstemmer/go-junit-report][jstemmer-link]

### Run Tests

```bash
go test ./...
```

[travis-badge]: https://travis-ci.org/jstemmer/go-junit-report.svg?branch=master
[travis-link]: https://travis-ci.org/jstemmer/go-junit-report
[report-badge]: https://goreportcard.com/badge/github.com/jstemmer/go-junit-report
[report-link]: https://goreportcard.com/report/github.com/jstemmer/go-junit-report
[jstemmer-link]: https://github.com/jstemmer/go-junit-report
