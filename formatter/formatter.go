package formatter

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/acarl005/stripansi"

	"github.com/hexon/go-junit-report/parser"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []JUnitTestSuite `xml:"testsuite"`
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Skipped    int             `xml:"skipped,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase `xml:"testcase"`
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Error       *JUnitError       `xml:"error,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
	SystemOut   string            `xml:",comment"` // A <system-out> element exists in <testsuite> but not in <testcase>
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitError contains data related to a test error.
type JUnitError struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

// JUnitReportXML writes a JUnit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func JUnitReportXML(report *parser.Report, noXMLHeader bool, goVersion string, fullPackageClassname bool, stripANSIEscape bool, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		ts := JUnitTestSuite{
			Tests:      len(pkg.Tests),
			Failures:   0,
			Errors:     0,
			Time:       formatTime(pkg.Duration),
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := pkg.Name
		if !fullPackageClassname {
			if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
				classname = pkg.Name[idx+1:]
			}
		}

		// properties
		if goVersion == "" {
			// if goVersion was not specified as a flag, fall back to version reported by runtime
			goVersion = runtime.Version()
		}
		ts.Properties = append(ts.Properties, JUnitProperty{"go.version", goVersion})
		if pkg.CoveragePct != "" {
			ts.Properties = append(ts.Properties, JUnitProperty{"coverage.statements.pct", pkg.CoveragePct})
		}

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      formatTime(test.Duration),
				Failure:   nil,
			}

			switch test.Result {
			case parser.SKIP:
				ts.Skipped++
				testCase.SkipMessage = &JUnitSkipMessage{
					Message: formatOutput(test.Output, stripANSIEscape),
				}
			case parser.ERROR:
				ts.Errors++
				testCase.Error = &JUnitError{
					Message:  "Error",
					Type:     "",
					Contents: formatOutput(test.Output, stripANSIEscape),
				}
			case parser.FAIL:
				ts.Failures++
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: formatOutput(test.Output, stripANSIEscape),
				}
			case parser.PASS:
				testCase.SystemOut = formatOutput(test.Output, stripANSIEscape)
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	if !noXMLHeader {
		writer.WriteString(xml.Header)
	}

	writer.Write(bytes)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}

func formatOutput(lines []string, stripANSIEscape bool) string {
	joined := strings.Join(lines, "\n")
	if stripANSIEscape {
		joined = stripansi.Strip(joined)
	}
	return joined
}
