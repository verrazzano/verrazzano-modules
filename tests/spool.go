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

type logMessage struct {
	Time    string  `json:"Time,omitempty"`
	Action  string  `json:"Action,omitempty"`
	Package string  `json:"Package,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

type writer struct {
	out                 io.Writer
	buf                 bytes.Buffer
	mtx                 *sync.Mutex
	lineCount           int
	spoolLogSummaryFile string
	stdout              io.Writer
}

func (w *writer) flush() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if len(w.buf.Bytes()) == 0 {
		return nil
	}
	w.clearLines()

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
func (w *writer) clearLines() {
	_, _ = fmt.Fprint(w.out, strings.Repeat(fmt.Sprintf("%c[%dA%c[2K", 27, 1, 27), w.lineCount))
}

func newWriter() *writer {
	w := &writer{
		mtx: &sync.Mutex{},
	}
	w.spoolLogSummaryFile = os.Getenv("SPOOL_LOG_SUMMARY")
	if w.spoolLogSummaryFile != "" {
		file, err := os.Create(w.spoolLogSummaryFile)
		if err != nil {
			handleError(err)
		}

		w.out = io.Writer(file)
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

		if logMessage.Test == "" && logMessage.Output != "" {
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

				if len(parts) == 1 && logMessage.Output != "" {
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

			if logMessage.Output != "" {
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
