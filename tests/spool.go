// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// logMessage represents the json record for a single log message.
type logMessage struct {
	Time    string  `json:"Time,omitempty"`
	Action  string  `json:"Action,omitempty"`
	Package string  `json:"Package,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// writer contains the buffer for writing formatted output and the Writers for
// stdout and formatted output. JSON log messages are read from a test spool log
// When spoolLogSummaryFile is defined, formatted and sorted messages are written to a sumamry file
// and raw formatted messages are routed to stdout. Otherwise only formatted and sorted
// messages are routed to stdout and no summary file is created.
type writer struct {
	out                 io.Writer
	buf                 bytes.Buffer
	mtx                 *sync.Mutex
	lineCount           int
	spoolLogSummaryFile string
	stdout              io.Writer
}

// flush rleases the contents of buffer to output.
// when spoolLogSummaryFile is defined, it is truncated on every flush so that it always contains sorter and formatted
// up-to-date messages.
// when spoolLogSummaryFile is noyt defined, it is assumed that the test is being run from a terminal and
// it attempts to replace the contents of terminal with latest sorted and formatted messages.
func (w *writer) flush() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if len(w.buf.Bytes()) == 0 {
		return nil
	}

	if w.spoolLogSummaryFile == "" {
		w.clearLines()
	} else {
		file, err := os.Create(w.spoolLogSummaryFile)
		if err != nil {
			handleError(err)
		}

		w.out = io.Writer(file)
		defer file.Close()
	}

	lines := 0
	var currentLine bytes.Buffer
	for _, b := range w.buf.Bytes() {
		if b == '\n' {
			lines++
			currentLine.Reset()
		} else {
			currentLine.Write([]byte{b})
		}
	}
	w.lineCount = lines

	_, err := w.out.Write(w.buf.Bytes())
	if err != nil {
		return err
	}

	w.buf.Reset()
	return nil

}

// clears the terminal window, equivalent of clear command.
func (w *writer) clearLines() {
	_, _ = fmt.Fprint(w.out, strings.Repeat(fmt.Sprintf("%c[%dA%c[2K", 27, 1, 27), w.lineCount))
}

func newWriter() *writer {
	w := &writer{
		mtx: &sync.Mutex{},
	}
	w.spoolLogSummaryFile = os.Getenv("SPOOL_LOG_SUMMARY")
	if w.spoolLogSummaryFile != "" {
		w.stdout = io.Writer(os.Stdout)
	} else {
		w.out = io.Writer(os.Stdout)
	}
	return w
}

func (w *writer) write(line string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	_, err := w.buf.Write([]byte(line))
	if err != nil {
		handleError(err)
	}
}

func (w *writer) print(line string) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	_, err := w.stdout.Write([]byte(line))
	if err != nil {
		handleError(err)
	}
}

// main reads the test log spool file that contains the log messages from test execution in json format.
// on each such message, it extracts information of package, suite, test and the actual log message and
// associates log messages to a test, suite or package and separately creates association between package and
// their suites, suites and its tests or package with tests without suites.
// When w.stdout is initialized, it prints the log messages on stdout as they are received in a contextual format i.e.
// <package>
//
//	<message> | <suite> | <test>
//					<message> | <test>
//									<message>
//
// Apart from that, it also sorts the packages, suites and tests based on their names and messages based on their time
// and summary information is sent to a buffer, which can further release that to a file or stdout. The representaion
// is depicted below:
// <package A, package paths are sorted lexicographically>
//
//	<any messages for package A, sorted by time>
//	<suite A of package A, suite are sorted lexicographically>
//		<any messages for suite A of package A, sorted by time>
//		<test A of suite A of package A, test names are sorted lexicographically>
//			<message m1 of test A received on time t1, messages are sorted based on time>
func main() {
	file, err := os.Open(os.Getenv("SPOOL_LOG"))
	if err != nil {
		handleError(err)
	}

	logMessagePackageMap := make(map[string]map[time.Time]string)
	logMessageSuiteMap := make(map[string]map[time.Time]string)
	logMessageTestMap := make(map[string]map[time.Time]string)
	packageSuiteMap := make(map[string]map[string]string)
	packageTestMap := make(map[string]map[string]string)
	suiteTestMap := make(map[string]map[string]string)
	w := newWriter()

	defer file.Close()
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			handleError(err)
		}

		if strings.Contains(line, "END SPOOL") {
			os.Exit(0)
			break
		}

		logMessage := &logMessage{}
		err = json.Unmarshal([]byte(line), logMessage)
		if err != nil {
			handleError(err)
		}

		if logMessage.Elapsed != float64(0) {
			continue
		}

		packageLogMessages, ok := logMessagePackageMap[logMessage.Package]
		if !ok {
			packageLogMessages = make(map[time.Time]string)
			logMessagePackageMap[logMessage.Package] = packageLogMessages
		}

		packageSuites, ok := packageSuiteMap[logMessage.Package]
		if !ok {
			packageSuites = make(map[string]string)
			packageSuiteMap[logMessage.Package] = packageSuites
		}

		packageTests, ok := packageTestMap[logMessage.Package]
		if !ok {
			packageTests = make(map[string]string)
			packageTestMap[logMessage.Package] = packageTests
		}

		if logMessage.Test == "" && logMessage.Output != "" && shouldBeLogged(logMessage.Output) {
			dateTime, err := time.Parse(time.RFC3339Nano, logMessage.Time)
			if err != nil {
				handleError(err)
			}

			packageLogMessages[dateTime] = logMessage.Output
			if w.stdout != nil {
				w.print(fmt.Sprintf("%s\n", logMessage.Package))
				w.print(fmt.Sprintf("\t%s: %s", dateTime.Format(time.RFC3339Nano), logMessage.Output))
			}
		}

		currentSuite := ""
		if logMessage.Test != "" {
			testName := logMessage.Test
			parts := strings.Split(testName, "/")
			if strings.Contains(parts[0], "Suite") {
				currentSuite = parts[0]
				packageSuites[currentSuite] = ""
				suiteLogMessages, ok := logMessageSuiteMap[currentSuite]
				if !ok {
					suiteLogMessages = make(map[time.Time]string)
					logMessageSuiteMap[currentSuite] = suiteLogMessages
				}

				suiteTests, ok := suiteTestMap[currentSuite]
				if !ok {
					suiteTests = make(map[string]string)
					suiteTestMap[currentSuite] = suiteTests
				}

				if len(parts) == 1 && logMessage.Output != "" && shouldBeLogged(logMessage.Output) {
					dateTime, err := time.Parse(time.RFC3339Nano, logMessage.Time)
					if err != nil {
						handleError(err)
					}

					suiteLogMessages[dateTime] = logMessage.Output
					if w.stdout != nil {
						w.print(fmt.Sprintf("%s\n", logMessage.Package))
						w.print(fmt.Sprintf("\t%s\n", currentSuite))
						w.print(fmt.Sprintf("\t\t%s: %s", dateTime.Format(time.RFC3339Nano), logMessage.Output))
					}
				}

				if len(parts) > 1 {
					parts = parts[1:]
					testName = strings.Join(parts, "/")
					suiteTests[testName] = ""
				}
			} else {
				packageTests[testName] = ""
			}

			testLogMessages, ok := logMessageTestMap[testName]
			if !ok {
				testLogMessages = make(map[time.Time]string)
				logMessageTestMap[testName] = testLogMessages
			}

			if logMessage.Output != "" && shouldBeLogged(logMessage.Output) {
				dateTime, err := time.Parse(time.RFC3339Nano, logMessage.Time)
				if err != nil {
					handleError(err)
				}

				testLogMessages[dateTime] = logMessage.Output
				if w.stdout != nil {
					if currentSuite != "" {
						w.print(fmt.Sprintf("%s\n", logMessage.Package))
						w.print(fmt.Sprintf("\t%s\n", currentSuite))
						w.print(fmt.Sprintf("\t\t%s\n", testName))
						w.print(fmt.Sprintf("\t\t\t%s: %s", dateTime.Format(time.RFC3339Nano), logMessage.Output))
					} else {
						w.print(fmt.Sprintf("%s\n", logMessage.Package))
						w.print(fmt.Sprintf("\t\t%s\n", testName))
						w.print(fmt.Sprintf("\t\t%s: %s", dateTime.Format(time.RFC3339Nano), logMessage.Output))
					}
				}
			}
		}

		packageNameKeys := make([]string, 0, len(logMessagePackageMap))
		for packageNameKey := range logMessagePackageMap {
			packageNameKeys = append(packageNameKeys, packageNameKey)
		}

		sort.Strings(packageNameKeys)
		for _, packageName := range packageNameKeys {
			w.write(fmt.Sprintf("%s\n", packageName))
			if packageSuites, ok := packageSuiteMap[packageName]; ok {
				suites := w.sortMapKeys(packageSuites)
				for _, suite := range suites {
					w.write(fmt.Sprintf("\t%s\n", suite))
					if suiteTests, ok := suiteTestMap[suite]; ok {
						tests := w.sortMapKeys(suiteTests)
						for _, test := range tests {
							w.write(fmt.Sprintf("\t\t%s\n", test))
							if testMessages, ok := logMessageTestMap[test]; ok {
								w.sortAndWriteMessages(testMessages, 3)
							}
						}
					}

					if suiteMessages, ok := logMessageSuiteMap[suite]; ok {
						w.sortAndWriteMessages(suiteMessages, 2)
					}
				}
			}

			if packageTests, ok := packageTestMap[packageName]; ok {
				tests := w.sortMapKeys(packageTests)
				for _, test := range tests {
					w.write(fmt.Sprintf("\t%s\n", test))
					if testMessages, ok := logMessageTestMap[test]; ok {
						w.sortAndWriteMessages(testMessages, 2)
					}
				}
			}
			w.sortAndWriteMessages(logMessagePackageMap[packageName], 1)
		}
		handleError(w.flush())
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("error while spooling logs, error: %v", err.Error())
		os.Exit(1)
	}
}

func (w *writer) sortAndWriteMessages(messages map[time.Time]string, numTabs int) {
	messageKeys := make([]time.Time, 0, len(messages))
	for messageKey := range messages {
		messageKeys = append(messageKeys, messageKey)

	}

	sort.Slice(messageKeys, func(i, j int) bool {
		return messageKeys[i].Before(messageKeys[j])
	})

	for _, timeKey := range messageKeys {
		w.write(fmt.Sprintf("%s%s: %s", strings.Repeat("\t", numTabs), timeKey.Format(time.RFC3339Nano), messages[timeKey]))
	}
}

func (w *writer) sortMapKeys(elements map[string]string) []string {
	keys := make([]string, 0, len(elements))
	for key := range elements {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func shouldBeLogged(message string) bool {
	return !(strings.HasPrefix(message, "=== PAUSE") || strings.HasPrefix(message, "=== CONT"))
}
