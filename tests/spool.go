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
	Time    string `json:"Time,omitempty"`
	Action  string `json:"Action,omitempty"`
	Package string `json:"Package,omitempty"`
	Test    string `json:"Test,omitempty"`
	Output  string `json:"Output,omitempty"`
	Elapsed string `json:"Elapsed,omitempty"`
}

type writer struct {
	Out io.Writer
	buf bytes.Buffer
	mtx *sync.Mutex
}

func (w *writer) flush() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	spoolLogFile := os.Getenv("SPOOL_LOG_FORMATTED")
	err := os.Truncate(spoolLogFile, int64(0))
	if err != nil {
		return fmt.Errorf("error while truncating spool log file %s", spoolLogFile)
	}

	file, err := os.OpenFile(spoolLogFile, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error while truncating spool log file %s", spoolLogFile)
	}

	defer file.Close()
	w.Out = io.Writer(file)

	if len(w.buf.Bytes()) == 0 {
		return nil
	}

	_, err = w.Out.Write(w.buf.Bytes())
	if err != nil {
		return err
	}

	w.buf.Reset()
	return nil

}

func newWriter() *writer {
	return &writer{
		mtx: &sync.Mutex{},
	}
}

func (w *writer) write(line string) (n int, err error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.buf.Write([]byte(line))
}

func main() {
	file, err := os.Open(os.Getenv("SPOOL_LOG"))
	logMessagePackageMap := make(map[string]map[time.Time]string)
	logMessageSuiteMap := make(map[string]map[time.Time]string)
	logMessageTestMap := make(map[string]map[time.Time]string)
	packageSuiteMap := make(map[string]map[string]string)
	packageTestMap := make(map[string]map[string]string)
	suiteTestMap := make(map[string]map[string]string)
	w := newWriter()

	if err != nil {
		return
	}
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
			file.Close()
			os.Exit(0)
		}

		logMessage := &logMessage{}
		err = json.Unmarshal([]byte(line), logMessage)
		if err != nil {
			handleError(err)
		}

		if logMessage.Elapsed != "" {
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
		}

		if logMessage.Test != "" {
			testName := logMessage.Test
			parts := strings.Split(testName, "/")
			if strings.Contains(parts[0], "Suite") {
				suite := parts[0]
				packageSuites[suite] = ""

				suiteLogMessages, ok := logMessageSuiteMap[suite]
				if !ok {
					suiteLogMessages = make(map[time.Time]string)
					logMessageSuiteMap[suite] = suiteLogMessages
				}

				suiteTests, ok := suiteTestMap[suite]
				if !ok {
					suiteTests = make(map[string]string)
					suiteTestMap[suite] = suiteTests
				}

				if len(parts) == 1 && logMessage.Output != "" {
					dateTime, err := time.Parse(time.RFC3339Nano, logMessage.Time)
					if err != nil {
						handleError(err)
					}

					suiteLogMessages[dateTime] = logMessage.Output
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
			}
		}

		packageNameKeys := make([]string, 0, len(logMessagePackageMap))
		for packageNameKey := range logMessagePackageMap {
			packageNameKeys = append(packageNameKeys, packageNameKey)
		}

		sort.Strings(packageNameKeys)
		for _, packageName := range packageNameKeys {
			_, err := w.write(fmt.Sprintf("%s\n", packageName))
			if err != nil {
				handleError(err)
			}
			if packageSuites, ok := packageSuiteMap[packageName]; ok {
				suites := make([]string, 0, len(packageSuites))
				for suiteNameKey := range packageSuites {
					suites = append(suites, suiteNameKey)
				}

				sort.Strings(suites)
				for _, suite := range suites {
					_, err := w.write(fmt.Sprintf("\t%s\n", suite))
					if err != nil {
						handleError(err)
					}

					if suiteTests, ok := suiteTestMap[suite]; ok {
						tests := make([]string, 0, len(suiteTests))
						for testNameKey := range suiteTests {
							tests = append(tests, testNameKey)
						}

						sort.Strings(tests)

						for _, test := range tests {
							_, err := w.write(fmt.Sprintf("\t\t:%s\n", test))
							if err != nil {
								handleError(err)
							}

							if testMessages, ok := logMessageTestMap[test]; ok {
								testKeys := make([]time.Time, 0, len(testMessages))
								for testKey := range testMessages {
									testKeys = append(testKeys, testKey)

								}

								sort.Slice(testKeys, func(i, j int) bool {
									return testKeys[i].Before(testKeys[j])
								})

								for _, timeKey := range testKeys {
									_, err := w.write(fmt.Sprintf("\t\t\t%s: %s", timeKey.Format(time.RFC3339Nano), testMessages[timeKey]))
									if err != nil {
										handleError(err)
									}
								}
							}
						}
					}

					if suiteMessages, ok := logMessageSuiteMap[suite]; ok {
						suiteKeys := make([]time.Time, 0, len(suiteMessages))
						for suiteKey := range suiteMessages {
							suiteKeys = append(suiteKeys, suiteKey)

						}

						sort.Slice(suiteKeys, func(i, j int) bool {
							return suiteKeys[i].Before(suiteKeys[j])
						})

						for _, timeKey := range suiteKeys {
							_, err := w.write(fmt.Sprintf("\t\t%s: %s", timeKey.Format(time.RFC3339Nano), suiteMessages[timeKey]))
							if err != nil {
								handleError(err)
							}
						}
					}
				}
			}

			packageMessages := logMessagePackageMap[packageName]
			packageKeys := make([]time.Time, 0, len(packageMessages))
			for packageKey := range packageMessages {
				packageKeys = append(packageKeys, packageKey)

			}

			sort.Slice(packageKeys, func(i, j int) bool {
				return packageKeys[i].Before(packageKeys[j])
			})

			for _, timeKey := range packageKeys {
				_, err := w.write(fmt.Sprintf("\t%s: %s", timeKey.Format(time.RFC3339Nano), packageMessages[timeKey]))
				if err != nil {
					handleError(err)
				}
			}
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
