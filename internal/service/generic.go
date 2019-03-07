package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/migotom/mt-bulk/internal/schema"
)

// HandlerFunc executes sequence of operations in context of service using established already connection and passed by context into handler.
type HandlerFunc func(context interface{}) error

// Service represents common interface for all supported services.
type Service interface {
	RunCmd(string) (string, error)
	GetDevice() *GenericDevice
}

// GenericDevice defines basic setup for device instance.
type GenericDevice struct {
	AppConfig          *schema.GeneralConfig
	Host               schema.Host
	currentPasswordIdx int
	matches            map[string]string
}

// ExecuteCommands executes provided list of commands using specified service.
func ExecuteCommands(ctx context.Context, d Service, commands []schema.Command) error {
	dev := d.GetDevice()
	for _, c := range commands {
		select {
		case <-ctx.Done():
			return nil
		default:
			for match, value := range dev.matches {
				fmt.Printf("checking match:%s (replace with value %s)\n", match, value)
				re := regexp.MustCompile(match)
				c.Body = re.ReplaceAllString(c.Body, value)
			}

			if dev.AppConfig.Verbose {
				log.Printf("[IP:%s]> %s\n", dev.Host.IP, c.Body)
			}

			var err error
			c.Result, err = d.RunCmd(c.Body)
			if err != nil {
				return err
			}

			if dev.AppConfig.Verbose {
				log.Printf("[IP:%s]< %s\n", dev.Host.IP, c.Result)
			}

			if c.Sleep.Duration > 0 {
				log.Printf("[IP:%s] (sleep %s)\n", dev.Host.IP, c.Sleep.Duration)
				time.Sleep(c.Sleep.Duration)
			}

			if c.Match != "" {
				re := regexp.MustCompile(c.Match)
				if matches := re.FindStringSubmatch(c.Result); len(matches) > 1 {
					for i := 1; i < len(matches); i++ {
						dev.matches[fmt.Sprintf("(%%{%s%d})", c.MatchPrefix, i)] = matches[i]
						fmt.Printf("found: key:%s value:%v\n", fmt.Sprintf("(%%{%s%d})", c.MatchPrefix, i), matches[i])
					}
				}
			}

			if c.SkipExpected {
				continue
			}

			if ok := strings.Contains(c.Result, c.Expect); !ok {
				return fmt.Errorf("Expected result to contain %s", c.Expect)
			}
		}
	}
	return nil
}
