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

	"github.com/jstemmer/go-junit-report/parser"
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
func JUnitReportXML(report *parser.Report, noXMLHeader bool, goVersion string, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		pkg.Benchmarks = mergeBenchmarks(pkg.Benchmarks)
		ts := JUnitTestSuite{
			Tests:      len(pkg.Tests) + len(pkg.Benchmarks),
			Failures:   0,
			Errors:     0,
			Time:       formatTime(pkg.Duration),
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
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
					Message: formatOutput(test.Output),
				}
			case parser.ERROR:
				ts.Errors++
				testCase.Error = &JUnitError{
					Message:  "Error",
					Type:     "",
					Contents: formatOutput(test.Output),
				}
			case parser.FAIL:
				ts.Failures++
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: formatOutput(test.Output),
				}
			case parser.PASS:
				testCase.SystemOut = formatOutput(test.Output)
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		// individual benchmarks
		for _, benchmark := range pkg.Benchmarks {
			benchmarkCase := JUnitTestCase{
				Classname: classname,
				Name:      benchmark.Name,
				Time:      formatBenchmarkTime(benchmark.Duration),
			}

			ts.TestCases = append(ts.TestCases, benchmarkCase)
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

func mergeBenchmarks(benchmarks []*parser.Benchmark) []*parser.Benchmark {
	var merged []*parser.Benchmark
	benchmap := make(map[string][]*parser.Benchmark)
	for _, bm := range benchmarks {
		if _, ok := benchmap[bm.Name]; !ok {
			merged = append(merged, &parser.Benchmark{Name: bm.Name})
		}
		benchmap[bm.Name] = append(benchmap[bm.Name], bm)
	}

	for _, bm := range merged {
		for _, b := range benchmap[bm.Name] {
			bm.Allocs += b.Allocs
			bm.Bytes += b.Bytes
			bm.Duration += b.Duration
		}
		n := len(benchmap[bm.Name])
		bm.Allocs /= n
		bm.Bytes /= n
		bm.Duration /= time.Duration(n)
	}

	return merged
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

func formatBenchmarkTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}

func formatOutput(lines []string) string {
	joined := strings.Join(lines, "\n")
	return stripansi.Strip(joined)
}
