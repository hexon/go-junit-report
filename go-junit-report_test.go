package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hexon/go-junit-report/formatter"
	"github.com/hexon/go-junit-report/parser"
)

type TestCase struct {
	name                 string
	reportName           string
	report               *parser.Report
	noXMLHeader          bool
	packageName          string
	fullPackageClassname bool
	stripANSIEscape      bool
}

var testCases = []TestCase{
	{
		name:       "01-pass.txt",
		reportName: "01-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestZ",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "02-fail.txt",
		reportName: "02-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 151 * time.Millisecond,
					Time:     151,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.FAIL,
							Output: []string{
								"file_test.go:11: Error message",
								"file_test.go:11: Longer",
								"\terror",
								"\tmessage.",
							},
						},
						{
							Name:     "TestTwo",
							Duration: 130 * time.Millisecond,
							Time:     130,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "03-skip.txt",
		reportName: "03-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 150 * time.Millisecond,
					Time:     150,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.SKIP,
							Output: []string{
								"file_test.go:11: Skip message",
							},
						},
						{
							Name:     "TestTwo",
							Duration: 130 * time.Millisecond,
							Time:     130,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "04-go_1_4.txt",
		reportName: "04-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "05-no_xml_header.txt",
		reportName: "05-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
		noXMLHeader: true,
	},
	{
		name:       "06-mixed.txt",
		reportName: "06-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name1",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
				{
					Name:     "package/name2",
					Duration: 151 * time.Millisecond,
					Time:     151,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.FAIL,
							Output: []string{
								"file_test.go:11: Error message",
								"file_test.go:11: Longer",
								"\terror",
								"\tmessage.",
							},
						},
						{
							Name:     "TestTwo",
							Duration: 130 * time.Millisecond,
							Time:     130,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
		noXMLHeader: true,
	},
	{
		name:       "07-compiled_test.txt",
		reportName: "07-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "test/package",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
		packageName: "test/package",
	},
	{
		name:       "08-parallel.txt",
		reportName: "08-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "github.com/dmitris/test-go-junit-report",
					Duration: 440 * time.Millisecond,
					Time:     440,
					Tests: []*parser.Test{
						{
							Name:     "TestDoFoo",
							Duration: 270 * time.Millisecond,
							Time:     270,
							Result:   parser.PASS,
							Output:   []string{"cov_test.go:10: DoFoo log 1", "cov_test.go:10: DoFoo log 2"},
						},
						{
							Name:     "TestDoFoo2",
							Duration: 160 * time.Millisecond,
							Time:     160,
							Result:   parser.PASS,
							Output:   []string{"cov_test.go:21: DoFoo2 log 1", "cov_test.go:21: DoFoo2 log 2"},
						},
					},
				},
			},
		},
	},
	{
		name:       "09-coverage.txt",
		reportName: "09-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestZ",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
					CoveragePct: "13.37",
				},
			},
		},
	},
	{
		name:       "10-multipkg-coverage.txt",
		reportName: "10-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package1/foo",
					Duration: 400 * time.Millisecond,
					Time:     400,
					Tests: []*parser.Test{
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestB",
							Duration: 300 * time.Millisecond,
							Time:     300,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
					CoveragePct: "10.0",
				},
				{
					Name:     "package2/bar",
					Duration: 4200 * time.Millisecond,
					Time:     4200,
					Tests: []*parser.Test{
						{
							Name:     "TestC",
							Duration: 4200 * time.Millisecond,
							Time:     4200,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
					CoveragePct: "99.8",
				},
			},
		},
	},
	{
		name:       "11-go_1_5.txt",
		reportName: "11-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 50 * time.Millisecond,
					Time:     50,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 30 * time.Millisecond,
							Time:     30,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "12-go_1_7.txt",
		reportName: "12-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 50 * time.Millisecond,
					Time:     50,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 10 * time.Millisecond,
							Time:     10,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestOne/Child",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestOne/Child#01",
							Duration: 30 * time.Millisecond,
							Time:     30,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestOne/Child=02",
							Duration: 40 * time.Millisecond,
							Time:     40,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo",
							Duration: 10 * time.Millisecond,
							Time:     10,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo/Child",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo/Child#01",
							Duration: 30 * time.Millisecond,
							Time:     30,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestTwo/Child=02",
							Duration: 40 * time.Millisecond,
							Time:     40,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestThree",
							Duration: 10 * time.Millisecond,
							Time:     10,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestThree/a#1",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestThree/a#1/b#1",
							Duration: 30 * time.Millisecond,
							Time:     30,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestThree/a#1/b#1/c#1",
							Duration: 40 * time.Millisecond,
							Time:     40,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestFour",
							Duration: 20 * time.Millisecond,
							Time:     20,
							Result:   parser.FAIL,
							Output:   []string{},
						},
						{
							Name:     "TestFour/#00",
							Duration: 0,
							Time:     0,
							Result:   parser.FAIL,
							Output: []string{
								"example.go:12: Expected abc  OBTAINED:",
								"	xyz",
								"example.go:123: Expected and obtained are different.",
							},
						},
						{
							Name:     "TestFour/#01",
							Duration: 0,
							Time:     0,
							Result:   parser.SKIP,
							Output: []string{
								"example.go:1234: Not supported yet.",
							},
						},
						{
							Name:     "TestFour/#02",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestFive",
							Duration: 0,
							Time:     0,
							Result:   parser.SKIP,
							Output: []string{
								"example.go:1392: Not supported yet.",
							},
						},
						{
							Name:     "TestSix",
							Duration: 0,
							Time:     0,
							Result:   parser.FAIL,
							Output: []string{
								"example.go:371: This should not fail!",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "13-syntax-error.txt",
		reportName: "13-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name/passing1",
					Duration: 100 * time.Millisecond,
					Time:     100,
					Tests: []*parser.Test{
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
				{
					Name:     "package/name/passing2",
					Duration: 100 * time.Millisecond,
					Time:     100,
					Tests: []*parser.Test{
						{
							Name:     "TestB",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
				{
					Name: "package/name/failing1",
					Tests: []*parser.Test{
						{
							Name:     "[build failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"failing1/failing_test.go:15: undefined: x",
							},
						},
					},
				},
				{
					Name: "package/name/failing2",
					Tests: []*parser.Test{
						{
							Name:     "[build failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"failing2/another_failing_test.go:20: undefined: y",
							},
						},
					},
				},
				{
					Name: "package/name/setupfailing1",
					Tests: []*parser.Test{
						{
							Name:     "[setup failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"setupfailing1/failing_test.go:4: cannot find package \"other/package\" in any of:",
								"\t/path/vendor (vendor tree)",
								"\t/path/go/root (from $GOROOT)",
								"\t/path/go/path (from $GOPATH)",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "14-panic.txt",
		reportName: "14-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/panic",
					Duration: 3 * time.Millisecond,
					Time:     3,
					Tests: []*parser.Test{
						{
							Name:   "Error",
							Result: parser.ERROR,
							Output: []string{
								"panic: init",
								"stacktrace",
							},
						},
					},
				},
				{
					Name:     "package/panic2",
					Duration: 3 * time.Millisecond,
					Time:     3,
					Tests: []*parser.Test{
						{
							Name:   "Error",
							Result: parser.ERROR,
							Output: []string{
								"panic: init",
								"stacktrace",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "15-empty.txt",
		reportName: "15-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/empty",
					Duration: 1 * time.Millisecond,
					Time:     1,
					Tests:    []*parser.Test{},
				},
			},
		},
	},
	{
		name:       "16-repeated-names.txt",
		reportName: "16-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/repeated-names",
					Duration: 1 * time.Millisecond,
					Time:     1,
					Tests: []*parser.Test{
						{
							Name:     "TestRepeat",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestRepeat",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestRepeat",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "17-race.txt",
		reportName: "17-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "race_test",
					Duration: 15 * time.Millisecond,
					Time:     15,
					Tests: []*parser.Test{
						{
							Name:     "TestRace",
							Duration: 0,
							Time:     0,
							Result:   parser.FAIL,
							Output: []string{
								"test output",
								"2 0xc4200153d0",
								"==================",
								"WARNING: DATA RACE",
								"Write at 0x00c4200153d0 by goroutine 7:",
								"  race_test.TestRace.func1()",
								"      race_test.go:13 +0x3b",
								"",
								"Previous write at 0x00c4200153d0 by goroutine 6:",
								"  race_test.TestRace()",
								"      race_test.go:15 +0x136",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"",
								"Goroutine 7 (running) created at:",
								"  race_test.TestRace()",
								"      race_test.go:14 +0x125",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"",
								"Goroutine 6 (running) created at:",
								"  testing.(*T).Run()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:697 +0x543",
								"  testing.runTests.func1()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:882 +0xaa",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"  testing.runTests()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:888 +0x4e0",
								"  testing.(*M).Run()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:822 +0x1c3",
								"  main.main()",
								"      _test/_testmain.go:52 +0x20f",
								"==================",
								"testing.go:610: race detected during execution of test",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "18-coverpkg.txt",
		reportName: "18-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package1/foo",
					Duration: 400 * time.Millisecond,
					Time:     400,
					Tests: []*parser.Test{
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestB",
							Duration: 300 * time.Millisecond,
							Time:     300,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
					CoveragePct: "10.0",
				},
				{
					Name:     "package2/bar",
					Duration: 4200 * time.Millisecond,
					Time:     4200,
					Tests: []*parser.Test{
						{
							Name:     "TestC",
							Duration: 4200 * time.Millisecond,
							Time:     4200,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
					CoveragePct: "99.8",
				},
			},
		},
	},
	{
		name:       "19-pass.txt",
		reportName: "19-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name",
					Duration: 160 * time.Millisecond,
					Time:     160,
					Tests: []*parser.Test{
						{
							Name:     "TestZ",
							Duration: 60 * time.Millisecond,
							Time:     60,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "20-parallel.txt",
		reportName: "20-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "pkg/parallel",
					Duration: 3010 * time.Millisecond,
					Time:     3010,
					Tests: []*parser.Test{
						{
							Name:     "FirstTest",
							Duration: 2 * time.Second,
							Time:     2000,
							Result:   parser.FAIL,
							Output: []string{
								"Message from first",
								"Supplemental from first",
								"parallel_test.go:14: FirstTest error",
							},
						},
						{
							Name:     "SecondTest",
							Duration: 1 * time.Second,
							Time:     1000,
							Result:   parser.FAIL,
							Output: []string{
								"Message from second",
								"parallel_test.go:23: SecondTest error",
							},
						},
						{
							Name:     "ThirdTest",
							Duration: 10 * time.Millisecond,
							Time:     10,
							Result:   parser.FAIL,
							Output: []string{
								"Message from third",
								"parallel_test.go:32: ThirdTest error",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "21-cached.txt",
		reportName: "21-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/one",
					Duration: 0,
					Time:     0,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "22-bench.txt",
		reportName: "22-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/basic",
					Duration: 3212 * time.Millisecond,
					Time:     3212,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkParse",
							Duration: 604 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkParse-8                     2000000\t       604 ns/op",
							},
						},
						{
							Name:     "BenchmarkReadingList",
							Duration: 1425 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkReadingList-8               1000000\t      1425 ns/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "23-benchmem.txt",
		reportName: "23-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/one",
					Duration: 9415 * time.Millisecond,
					Time:     9415,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkIpsHistoryInsert",
							Duration: 52568 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkIpsHistoryInsert-8 30000\t52568 ns/op\t24879 B/op\t494 allocs/op",
							},
						},
						{
							Name:     "BenchmarkIpsHistoryLookup",
							Duration: 15208 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkIpsHistoryLookup-8 100000\t15208 ns/op\t7369 B/op\t143 allocs/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "24-benchtests.txt",
		reportName: "24-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package3/baz",
					Duration: 1382 * time.Millisecond,
					Time:     1382,
					Tests: []*parser.Test{
						{
							Name:     "TestNew",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestNew/no",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestNew/normal",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "TestWriteThis",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "BenchmarkDeepMerge",
							Duration: 2611 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkDeepMerge-8      500000       2611 ns/op     1110 B/op       16 allocs/op",
							},
						},
						{
							Name:     "BenchmarkNext",
							Duration: 100 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkNext-8           500000       100 ns/op      100 B/op        1 allocs/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "25-benchcount.txt",
		reportName: "25-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "pkg/count",
					Duration: 14211 * time.Millisecond,
					Time:     14211,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkNew",
							Duration: 345 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkNew-8   \t 5000000\t       350 ns/op\t      80 B/op\t       3 allocs/op",
								"BenchmarkNew-8   \t 5000000\t       357 ns/op\t      80 B/op\t       3 allocs/op",
								"BenchmarkNew-8   \t 5000000\t       354 ns/op\t      80 B/op\t       3 allocs/op",
								"BenchmarkNew-8   \t 5000000\t       358 ns/op\t      80 B/op\t       3 allocs/op",
								"BenchmarkNew-8   \t 5000000\t       345 ns/op\t      80 B/op\t       3 allocs/op",
							},
						},
						{
							Name:     "BenchmarkFew",
							Duration: 102 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkFew-8   \t 5000000\t       100 ns/op\t      20 B/op\t       1 allocs/op",
								"BenchmarkFew-8   \t 5000000\t       105 ns/op\t      20 B/op\t       1 allocs/op",
								"BenchmarkFew-8   \t 5000000\t       102 ns/op\t      20 B/op\t       1 allocs/op",
								"BenchmarkFew-8   \t 5000000\t       102 ns/op\t      20 B/op\t       1 allocs/op",
								"BenchmarkFew-8   \t 5000000\t       102 ns/op\t      20 B/op\t       1 allocs/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "26-testbenchmultiple.txt",
		reportName: "26-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "mycode/common",
					Duration: 7267 * time.Millisecond,
					Time:     7267,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkParse",
							Duration: 1591 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkParse-8                   \t 1000000\t      1591 ns/op",
							},
						},
						{
							Name:     "BenchmarkNewTask",
							Duration: 391 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkNewTask-8                 \t 3000000\t       391 ns/op",
							},
						},
					},
				},
				{
					Name:     "mycode/benchmarks/channels",
					Duration: 47084 * time.Millisecond,
					Time:     47084,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkFanout/Channel/10",
							Duration: 4673 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkFanout/Channel/10-8         \t  500000\t      4673 ns/op",
							},
						},
						{
							Name:     "BenchmarkFanout/Channel/100",
							Duration: 24965 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkFanout/Channel/100-8        \t   50000\t     24965 ns/op",
							},
						},
						{
							Name:     "BenchmarkFanout/Channel/1000",
							Duration: 195672 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkFanout/Channel/1000-8       \t   10000\t    195672 ns/op",
							},
						},
						{
							Name:     "BenchmarkFanout/Channel/10000",
							Duration: 2410200 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkFanout/Channel/10000-8      \t     500\t   2410200 ns/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "27-benchdecimal.txt",
		reportName: "27-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "really/small",
					Duration: 4344 * time.Millisecond,
					Time:     4344,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkItsy",
							Duration: 45 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkItsy-8    \t  30000000\t         45.7 ns/op",
							},
						},
						{
							Name:     "BenchmarkTeeny",
							Duration: 2 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkTeeny-8      1000000000\t         2.12 ns/op",
							},
						},
						{
							Name:     "BenchmarkWeeny",
							Duration: 0 * time.Second,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkWeeny-8      2000000000\t         0.26 ns/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "28-bench-1cpu.txt",
		reportName: "28-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "single/cpu",
					Duration: 9467 * time.Millisecond,
					Time:     9467,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkRing",
							Duration: 74 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkRing        \t20000000\t        74.2 ns/op",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "29-bench-16cpu.txt",
		reportName: "29-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "sixteen/cpu",
					Duration: 1522 * time.Millisecond,
					Time:     1522,
					Tests: []*parser.Test{
						{
							Name:     "BenchmarkRingaround",
							Duration: 13571 * time.Nanosecond,
							Result:   parser.PASS,
							Output: []string{
								"BenchmarkRingaround-16    \t  100000\t     13571 ns/op",
							},
						},
					},
				},
			},
		},
	},
	{
		// generated by running go test on https://gist.github.com/liggitt/09a021ccec988b19917e0c2d60a18ee9
		name:       "30-stdout.txt",
		reportName: "30-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name1",
					Duration: 4567 * time.Millisecond,
					Time:     4567,
					Tests: []*parser.Test{
						{
							Name:     "TestFailWithStdoutAndTestOutput",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.FAIL,
							Output: []string{
								`multi`,
								`line`,
								`stdout`,
								`single-line stdout`,
								`example_test.go:13: single-line error`,
								`example_test.go:14: multi`,
								`    line`,
								`    error`,
							},
						},
						{
							Name:     "TestFailWithStdoutAndNoTestOutput",
							Duration: 150 * time.Millisecond,
							Time:     150,
							Result:   parser.FAIL,
							Output: []string{
								`multi`,
								`line`,
								`stdout`,
								`single-line stdout`,
							},
						},
						{
							Name:     "TestFailWithTestOutput",
							Duration: 200 * time.Millisecond,
							Time:     200,
							Result:   parser.FAIL,
							Output: []string{
								`example_test.go:26: single-line error`,
								`example_test.go:27: multi`,
								`    line`,
								`    error`,
							},
						},
						{
							Name:     "TestFailWithNoTestOutput",
							Duration: 250 * time.Millisecond,
							Time:     250,
							Result:   parser.FAIL,
							Output:   []string{},
						},

						{
							Name:     "TestPassWithStdoutAndTestOutput",
							Duration: 300 * time.Millisecond,
							Time:     300,
							Result:   parser.PASS,
							Output: []string{
								`multi`,
								`line`,
								`stdout`,
								`single-line stdout`,
								`example_test.go:39: single-line info`,
								`example_test.go:40: multi`,
								`    line`,
								`    info`,
							},
						},
						{
							Name:     "TestPassWithStdoutAndNoTestOutput",
							Duration: 350 * time.Millisecond,
							Time:     350,
							Result:   parser.PASS,
							Output: []string{
								`multi`,
								`line`,
								`stdout`,
								`single-line stdout`,
							},
						},
						{
							Name:     "TestPassWithTestOutput",
							Duration: 400 * time.Millisecond,
							Time:     400,
							Result:   parser.PASS,
							Output: []string{
								`example_test.go:51: single-line info`,
								`example_test.go:52: multi`,
								`    line`,
								`    info`,
							},
						},
						{
							Name:     "TestPassWithNoTestOutput",
							Duration: 500 * time.Millisecond,
							Time:     500,
							Result:   parser.PASS,
							Output:   []string{},
						},

						{
							Name:     "TestSubtests",
							Duration: 2270 * time.Millisecond,
							Time:     2270,
							Result:   parser.FAIL,
						},
						{
							Name:     "TestSubtests/TestFailWithStdoutAndTestOutput",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.FAIL,
							Output: []string{
								`1 multi`,
								`line`,
								`stdout`,
								`1 single-line stdout`,
								`example_test.go:65: 1 single-line error`,
								`example_test.go:66: 1 multi`,
								`    line`,
								`    error`,
							},
						},
						{
							Name:     "TestSubtests/TestFailWithStdoutAndNoTestOutput",
							Duration: 150 * time.Millisecond,
							Time:     150,
							Result:   parser.FAIL,
							Output: []string{
								`2 multi`,
								`line`,
								`stdout`,
								`2 single-line stdout`,
							},
						},
						{
							Name:     "TestSubtests/TestFailWithTestOutput",
							Duration: 200 * time.Millisecond,
							Time:     200,
							Result:   parser.FAIL,
							Output: []string{
								`example_test.go:78: 3 single-line error`,
								`example_test.go:79: 3 multi`,
								`    line`,
								`    error`,
							},
						},
						{
							Name:     "TestSubtests/TestFailWithNoTestOutput",
							Duration: 250 * time.Millisecond,
							Time:     250,
							Result:   parser.FAIL,
							Output:   []string{},
						},

						{
							Name:     "TestSubtests/TestPassWithStdoutAndTestOutput",
							Duration: 300 * time.Millisecond,
							Time:     300,
							Result:   parser.PASS,
							Output: []string{
								`4 multi`,
								`line`,
								`stdout`,
								`4 single-line stdout`,
								`example_test.go:91: 4 single-line info`,
								`example_test.go:92: 4 multi`,
								`    line`,
								`    info`,
							},
						},
						{
							Name:     "TestSubtests/TestPassWithStdoutAndNoTestOutput",
							Duration: 350 * time.Millisecond,
							Time:     350,
							Result:   parser.PASS,
							Output: []string{
								`5 multi`,
								`line`,
								`stdout`,
								`5 single-line stdout`,
							},
						},
						{
							Name:     "TestSubtests/TestPassWithTestOutput",
							Duration: 400 * time.Millisecond,
							Time:     400,
							Result:   parser.PASS,
							Output: []string{
								`example_test.go:103: 6 single-line info`,
								`example_test.go:104: 6 multi`,
								`    line`,
								`    info`,
							},
						},
						{
							Name:     "TestSubtests/TestPassWithNoTestOutput",
							Duration: 500 * time.Millisecond,
							Time:     500,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "31-syntax-error-test-binary.txt",
		reportName: "31-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "package/name/passing1",
					Duration: 100 * time.Millisecond,
					Time:     100,
					Tests: []*parser.Test{
						{
							Name:     "TestA",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
				{
					Name:     "package/name/passing2",
					Duration: 100 * time.Millisecond,
					Time:     100,
					Tests: []*parser.Test{
						{
							Name:     "TestB",
							Duration: 100 * time.Millisecond,
							Time:     100,
							Result:   parser.PASS,
							Output:   []string{},
						},
					},
				},
				{
					Name: "package/name/failing1",
					Tests: []*parser.Test{
						{
							Name:     "[build failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"failing1/failing_test.go:15: undefined: x",
							},
						},
					},
				},
				{
					Name: "package/name/failing2",
					Tests: []*parser.Test{
						{
							Name:     "[build failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"failing2/another_failing_test.go:20: undefined: y",
							},
						},
					},
				},
				{
					Name: "package/name/setupfailing1",
					Tests: []*parser.Test{
						{
							Name:     "[setup failed]",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output: []string{
								"setupfailing1/failing_test.go:4: cannot find package \"other/package\" in any of:",
								"\t/path/vendor (vendor tree)",
								"\t/path/go/root (from $GOROOT)",
								"\t/path/go/path (from $GOPATH)",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "32-failed-summary.txt",
		reportName: "32-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "github.com/jstemmer/test/failedsummary",
					Duration: 5 * time.Millisecond,
					Time:     5,
					Tests: []*parser.Test{
						{
							Name:     "TestOne",
							Duration: 0,
							Time:     0,
							Result:   parser.PASS,
							Output:   []string{},
						},
						{
							Name:     "Error",
							Duration: 0,
							Time:     0,
							Result:   parser.ERROR,
							Output:   []string{"panic: panic"},
						},
					},
				},
			},
		},
	},
	{
		name:       "33-fail-go_1_15.txt",
		reportName: "33-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:     "example.com/test",
					Duration: 5 * time.Millisecond,
					Time:     5,
					Tests: []*parser.Test{
						{
							Name:     "TestFoo",
							Duration: 0,
							Time:     0,
							Result:   parser.FAIL,
							Output: []string{
								"foo_test.go:6: Error message",
							},
						},
						{
							Name:     "TestBar",
							Duration: 0,
							Time:     0,
							Result:   parser.FAIL,
							Output: []string{
								"foo_test.go:10: Longer",
								"    error",
								"    message.",
							},
						},
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := os.Open("testdata/" + testCase.name)
			if err != nil {
				t.Fatal(err)
			}

			report, err := parser.Parse(file, testCase.packageName)
			if err != nil {
				t.Fatalf("error parsing: %s", err)
			}

			if report == nil {
				t.Fatalf("Report == nil")
			}

			expected := testCase.report
			if len(report.Packages) != len(expected.Packages) {
				t.Fatalf("Report packages == %d, want %d", len(report.Packages), len(expected.Packages))
			}

			for i, pkg := range report.Packages {
				expPkg := expected.Packages[i]

				t.Run(expPkg.Name, func(t *testing.T) {
					if pkg.Name != expPkg.Name {
						t.Errorf("Package.Name == %s, want %s", pkg.Name, expPkg.Name)
					}

					if pkg.Duration != expPkg.Duration {
						t.Errorf("Package.Duration == %s, want %s", pkg.Duration, expPkg.Duration)
					}

					// pkg.Time is deprecated
					if pkg.Time != expPkg.Time {
						t.Errorf("Package.Time == %d, want %d", pkg.Time, expPkg.Time)
					}

					if len(pkg.Tests) != len(expPkg.Tests) {
						t.Fatalf("Package Tests == %d, want %d", len(pkg.Tests), len(expPkg.Tests))
					}

					for j, test := range pkg.Tests {
						expTest := expPkg.Tests[j]

						t.Run(expTest.Name, func(t *testing.T) {
							if test.Name != expTest.Name {
								t.Errorf("Test.Name == %s, want %s", test.Name, expTest.Name)
							}

							if test.Duration != expTest.Duration {
								t.Errorf("Test.Duration == %s, want %s", test.Duration, expTest.Duration)
							}

							// test.Time is deprecated
							if test.Time != expTest.Time {
								t.Errorf("Test.Time == %d, want %d", test.Time, expTest.Time)
							}

							if test.Result != expTest.Result {
								t.Errorf("Test.Result == %v, want %v", test.Result, expTest.Result)
							}

							testOutput := strings.Join(test.Output, "\n")
							expTestOutput := strings.Join(expTest.Output, "\n")
							if testOutput != expTestOutput {
								t.Errorf("Test.Output\nEXP: %q\nGOT: %q", expTestOutput, testOutput)
							}
						})
					}

					if pkg.CoveragePct != expPkg.CoveragePct {
						t.Errorf("Package.CoveragePct == %s, want %s", pkg.CoveragePct, expPkg.CoveragePct)
					}
				})
			}
		})
	}
}

func TestJUnitFormatter(t *testing.T) {
	testJUnitFormatter(t, "")
}

func TestVersionFlag(t *testing.T) {
	testJUnitFormatter(t, "custom-version")
}

func testJUnitFormatter(t *testing.T, goVersion string) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			report, err := loadTestReport(testCase.reportName, goVersion)
			if err != nil {
				t.Fatal(err)
			}

			var junitReport bytes.Buffer

			if err = formatter.JUnitReportXML(testCase.report, testCase.noXMLHeader, goVersion, testCase.stripANSIEscape, testCase.fullPackageClassname, &junitReport); err != nil {
				t.Fatal(err)
			}

			if string(junitReport.Bytes()) != report {
				t.Errorf("Report XML\nEXP:\n%s\nGOT:\n%s", report, string(junitReport.Bytes()))
			}
		})
	}
}

func loadTestReport(name, goVersion string) (string, error) {
	contents, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
		return "", err
	}

	if goVersion == "" {
		// if goVersion is not specified, default to runtime version
		goVersion = runtime.Version()
	}

	// replace value="1.0" With actual version
	report := strings.Replace(string(contents), `value="1.0"`, fmt.Sprintf(`value="%s"`, goVersion), -1)

	return report, nil
}
