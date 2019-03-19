package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/migotom/mt-bulk/internal/schema"
)

// HandlerFunc executes sequence of operations in context of service using established already connection and passed by context into handler.
type HandlerFunc func(service Service) error

// ApplicationStatus stores final execution status that should be returned to OS.
type ApplicationStatus struct {
	code int
	sync.Mutex
}

// SetCode sets application status code.
func (app *ApplicationStatus) SetCode(status int) {
	app.Lock()
	defer app.Unlock()
	app.code = status
}

// Get returns application status code.
func (app *ApplicationStatus) Get() int {
	app.Lock()
	defer app.Unlock()
	return app.code
}

// Service represents common interface for all supported services.
type Service interface {
	GetUser() string
	GetPasswords() []string
	GetPort() string
	GetDevice() *GenericDevice

	SetHost(schema.Host)
	SetConfig(*schema.GeneralConfig)

	HandleSequence(ctx context.Context, handler HandlerFunc) error
	RunCmd(string, *regexp.Regexp) (string, error)

	Close() error
}

type CopyFiler interface {
	CopyFile(local, remote string) error
}

// GenericDevice defines basic setup for device instance.
type GenericDevice struct {
	AppConfig          *schema.GeneralConfig
	Host               schema.Host
	currentPasswordIdx int
	matches            map[string]string
}

func (d *GenericDevice) SetHost(host schema.Host) {
	d.Host = host
}

func (d *GenericDevice) SetConfig(config *schema.GeneralConfig) {
	d.AppConfig = config
}

// expect reads bytes from reader and waits for timeout or expectec regexp value
func (d *GenericDevice) expect(reader io.Reader, expect *regexp.Regexp) (result string, err error) {
	resultChan := make(chan string)
	errorChan := make(chan struct{})

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
			parsedResponse := regexp.MustCompile("\r").ReplaceAllString(s.String(), "")

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
		case <-time.After(10 * time.Second):
			err = fmt.Errorf("timeout of waiting to expected result %s", expect.String())
			return
		}
	}
}

// ExecuteCommands executes provided list of commands using specified service.
func ExecuteCommands(ctx context.Context, d Service, commands []schema.Command) (err error) {
	dev := d.GetDevice()

	for _, c := range commands {
		select {
		case <-ctx.Done():
			return nil
		default:
			for match, value := range dev.matches {
				fmt.Printf("checking match:%s (replace with value %s)\n", match, value)
				c.Body = regexp.MustCompile(match).ReplaceAllString(c.Body, value)
			}

			if dev.AppConfig.Verbose {
				log.Printf("[IP:%s] < %s\n", dev.Host.IP, c.Body)
			}

			var expect *regexp.Regexp
			if c.Expect != "" {
				expect = regexp.MustCompile(c.Expect)
			}

			if c.Result, err = d.RunCmd(c.Body, expect); err != nil {
				return fmt.Errorf("command processing error: %s", err)
			}

			if dev.AppConfig.Verbose {
				log.Printf("[IP:%s] > %s\n", dev.Host.IP, c.Result)
			}

			if c.Sleep.Duration > 0 {
				log.Printf("[IP:%s] (sleep %s)\n", dev.Host.IP, c.Sleep.Duration)
				time.Sleep(c.Sleep.Duration)
			}

			if matches := regexp.MustCompile(c.Match).FindStringSubmatch(c.Result); len(matches) > 1 {
				for i := 1; i < len(matches); i++ {
					dev.matches[fmt.Sprintf("(%%{%s%d})", c.MatchPrefix, i)] = matches[i]
					fmt.Printf("found: key:%s value:%v\n", fmt.Sprintf("(%%{%s%d})", c.MatchPrefix, i), matches[i])
				}
			}
		}
	}
	return nil
}
