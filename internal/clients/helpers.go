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
	errorChan := make(chan struct{})

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
				errorChan <- struct{}{}
				return
			}

			s.WriteString(string(buf[:byteCount]))
			parsedResponse := newLineRx.ReplaceAllString(s.String(), "")

			if expect.MatchString(parsedResponse) {
				resultChan <- parsedResponse
				break
			}
		}
	}()

	for {
		select {
		case result = <-resultChan:
			return
		case <-errorChan:
			err = fmt.Errorf("expected result %s", expect.String())
			return
		case <-time.After(3 * time.Second):
			err = fmt.Errorf("timeout of waiting to expected result %s", expect.String())
			return
		}
	}
}
