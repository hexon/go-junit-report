package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hexon/go-junit-report/formatter"
	"github.com/hexon/go-junit-report/parser"
)

var (
	noXMLHeader          = flag.Bool("no-xml-header", false, "do not print xml header")
	packageName          = flag.String("package-name", "", "specify a package name (compiled test have no package name in output)")
	goVersionFlag        = flag.String("go-version", "", "specify the value to use for the go.version property in the generated XML")
	setExitCode          = flag.Bool("set-exit-code", false, "set exit code to 1 if tests failed")
	stripANSIEscape      = flag.Bool("strip-ansi-escape-codes", false, "strip ANSI escape codes (terminal color codes)")
	fullPackageClassname = flag.Bool("full-package-classname", false, "use the full package name as the test classname instead of just the last part")
)

func main() {
	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "%s does not accept positional arguments\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	// Read input
	report, err := parser.Parse(os.Stdin, *packageName)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Write xml
	err = formatter.JUnitReportXML(report, *noXMLHeader, *goVersionFlag, *fullPackageClassname, *stripANSIEscape, os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}

	if *setExitCode && report.Failures() > 0 {
		os.Exit(1)
	}
}
