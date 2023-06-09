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
	"strings"
	"sync"
	"time"
)

type LogMessage struct {
	Time    string `json:"Time,omitempty"`
	Action  string `json:"Action,omitempty"`
	Package string `json:"Package,omitempty"`
	Test    string `json:"Test,omitempty"`
	Output  string `json:"Output,omitempty"`
	Elapsed string `json:"Elapsed,omitempty"`
	Suite   string `json:"Suite,omitempty"`
}

type Writer struct {
	Out       io.Writer
	buf       bytes.Buffer
	mtx       *sync.Mutex
	lineCount int
}

const ESC = 27

var clear = fmt.Sprintf("%c[%dA%c[2K", ESC, 1, ESC)

func (w *Writer) clearLines() {
	_, _ = fmt.Fprint(w.Out, strings.Repeat(clear, w.lineCount))
}

func (w *Writer) Flush() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	if len(w.buf.Bytes()) == 0 {
		return nil
	}
	w.clearLines()

	lines := 0
	for _, b := range w.buf.Bytes() {
		if b == '\n' {
			lines++
		}
	}
	w.lineCount = lines
	_, err := w.Out.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

func New() *Writer {
	return &Writer{
		Out: io.Writer(os.Stdout),
		mtx: &sync.Mutex{},
	}
}

// Write save the contents of buf to the writer b. The only errors returned are ones encountered while writing to the underlying buffer.
func (w *Writer) Write(line string) (n int, err error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.buf.Write([]byte(line + "\n"))
}

func main() {
	file, err := os.Open(os.Getenv("SPOOL_LOG"))
	logMessagePackageMap := make(map[string]map[string]string)
	logMessageSuiteMap := make(map[string]map[string]string)
	logMessageTestMap := make(map[string]map[string]string)
	packageSuiteMap := make(map[string]map[string]string)
	packageTestMap := make(map[string]map[string]string)
	suiteTestMap := make(map[string]map[string]string)
	w := New()

	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// without this sleep you would hogg the CPU
				time.Sleep(500 * time.Millisecond)
				// truncated ?
				truncated, errTruncated := isTruncated(file)
				if errTruncated != nil {
					break
				}
				if truncated {
					// seek from start
					_, errSeekStart := file.Seek(0, io.SeekStart)
					if errSeekStart != nil {
						break
					}
				}
				continue
			}
			break
		}
		if strings.Contains(line, "END SPOOL") {
			os.Exit(0)
		}

		logMessage := &LogMessage{}
		json.Unmarshal([]byte(line), logMessage)
		if logMessage.Elapsed != "" {
			continue
		}

		packageLogMessages, ok := logMessagePackageMap[logMessage.Package]
		if !ok {
			packageLogMessages = make(map[string]string)
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
			packageLogMessages[logMessage.Time] = logMessage.Output
			continue
		}

		if logMessage.Test != "" {
			testName := logMessage.Test
			parts := strings.Split(testName, "/")
			if strings.Contains(parts[0], "Suite") {
				suite := parts[0]
				packageSuites[suite] = ""

				suiteLogMessages, ok := logMessageSuiteMap[suite]
				if !ok {
					suiteLogMessages = make(map[string]string)
					logMessageSuiteMap[suite] = suiteLogMessages
				}

				suiteTests, ok := suiteTestMap[suite]
				if !ok {
					suiteTests = make(map[string]string)
					suiteTestMap[suite] = suiteTests
				}

				if len(parts) == 1 && logMessage.Output != "" {
					suiteLogMessages[logMessage.Time] = logMessage.Output
					continue
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
				testLogMessages = make(map[string]string)
				logMessageTestMap[testName] = testLogMessages
			}

			if logMessage.Output != "" {
				testLogMessages[logMessage.Time] = logMessage.Output
			}
		}

		for packageName, packageMessages := range logMessagePackageMap {
			w.Write(packageName)
			if suites, ok := packageSuiteMap[packageName]; ok {
				for suite := range suites {
					w.Write("\t" + suite)
					if tests, ok := suiteTestMap[suite]; ok {
						for test := range tests {
							w.Write("\t\t" + test)
							if testMessages, ok := logMessageTestMap[test]; ok {
								for testTime, testMessage := range testMessages {
									w.Write("\t\t\t" + testTime + ": " + testMessage)
								}
							}
						}
					}
					if suiteMessages, ok := logMessageSuiteMap[suite]; ok {
						for suiteTime, suiteMessage := range suiteMessages {
							w.Write("\t\t" + suiteTime + ": " + suiteMessage)
						}
					}
				}
			}
			for packageTime, packageMessage := range packageMessages {
				w.Write("\t" + packageTime + ": " + packageMessage)
			}
		}
		w.Flush()
	}
}

func isTruncated(file *os.File) (bool, error) {
	// current read position in a file
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	// file stat to get the size
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fileInfo.Size(), nil
}
