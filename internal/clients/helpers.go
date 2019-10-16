package clients

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

func waitForExpected(reader io.Reader, expect *regexp.Regexp) (result string, err error) {
	resultChan := make(chan string)
	errorChan := make(chan error)

	newLineRx := regexp.MustCompile("\r")

	go func() {
		defer close(resultChan)
		defer close(errorChan)

		var s strings.Builder
		for {
			time.Sleep(time.Millisecond * 100)

			buf := make([]byte, 1024*10)
			byteCount, err := reader.Read(buf)
			if err != nil {
				errorChan <- err
				return
			}
			s.WriteString(string(buf[:byteCount]))
			parsedResponse := newLineRx.ReplaceAllString(s.String(), "")
			resultChan <- parsedResponse

			if expect.MatchString(parsedResponse) {
				errorChan <- nil
				break
			}
		}
	}()

	for {
		select {
		case result = <-resultChan:
		case err = <-errorChan:
			return
		case <-time.After(3 * time.Second):
			err = fmt.Errorf("timeout on waiting to expected result: %s", expect.String())
			return
		}
	}
}
