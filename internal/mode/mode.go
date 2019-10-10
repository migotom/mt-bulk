package mode

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
)

// EstablishConnection tries to establish connection for provided host by specified client.
// It tries to connect by retries number and list of passwords defined in client's configuration.
func EstablishConnection(ctx context.Context, client clients.Client, host *entities.Host) (err error) {
	err = errors.New("unexpected error during connection establishing")

	defer func() {
		if err != nil {
			err = fmt.Errorf("could not be able to establish connection (%s) (%v)", host, err)
		}
	}()

	config := client.GetConfig()
	host.SetDefaults(config.DefaultPort, config.DefaultUser, config.DefaultPassword)
	for retry := 0; retry < config.Retries; retry++ {
		for idx, password := range host.GetPasswords() {

			select {
			case <-ctx.Done():
				return errors.New("context done")

			default:
				fmt.Printf("%s -> establishing connection, password #%d (retry #%d)\n", host, idx, retry)

				err = client.Connect(ctx, host.IP, host.Port, host.User, password)
				if err != nil {
					if _, ok := err.(clients.ErrorWrongPassword); ok {
						continue
					}
					if _, ok := err.(clients.ErrorRetryable); ok {
						break
					}
				}
				return
			}

			// store valid password for this device
			host.Password = password
		}
	}
	return
}

// ExecuteCommands executes provided list of commands using specified client.
func ExecuteCommands(ctx context.Context, d clients.Client, commands []entities.Command) ([]string, error) {
	allMatches := make(map[string]string)
	run := func(c entities.Command, responseChan chan<- string, errChan chan<- error) {
		defer close(responseChan)
		defer close(errChan)

		for match, value := range allMatches {
			c.Body = regexp.MustCompile(match).ReplaceAllString(c.Body, value)
		}

		var expect *regexp.Regexp
		if c.Expect != "" {
			expect = regexp.MustCompile(c.Expect)
		}

		result, err := d.RunCmd(c.Body, expect)
		if err != nil {
			errChan <- fmt.Errorf("command processing error: %s", err)
			return
		}

		if c.SleepMs > 0 {
			time.Sleep(time.Duration(c.SleepMs) * time.Millisecond)
		}

		responseChan <- result

		if commandMatches := regexp.MustCompile(c.Match).FindStringSubmatch(result); len(commandMatches) > 1 {
			for i := 1; i < len(commandMatches); i++ {
				allMatches[fmt.Sprintf("(%%{%s%d})", c.MatchPrefix, i)] = commandMatches[i]
				responseChan <- fmt.Sprintf("/<mt-bulk:regexp> \"%s\" set key \"%s\" with value %v", c.Match, fmt.Sprintf("%%{%s%d}", c.MatchPrefix, i), commandMatches[i])
			}
		}
	}

	executed := make([]string, 0, len(commands))
	for _, c := range commands {
		errChan := make(chan error)
		responseChan := make(chan string)

		go run(c, responseChan, errChan)

	commandParseLoop:
		for {
			select {
			case <-ctx.Done():
				return executed, nil
			case <-time.After(30 * time.Second):
				return executed, fmt.Errorf("ExecuteCommands timeouted")
			case err, ok := <-errChan:
				if !ok {
					break commandParseLoop
				}
				return executed, fmt.Errorf("%v (%s)", err, c)
			case response, ok := <-responseChan:
				if !ok {
					break commandParseLoop
				}
				executed = append(executed, response)
			}
		}
	}
	return executed, nil
}

// OperationModeFunc represents operation mode function.
type OperationModeFunc func(context.Context, clients.Client, *entities.Job) ([]string, error)

const (
	// ChangePasswordMode is change password operation name.
	ChangePasswordMode = "ChangePassword"
	// CustomSSHMode is custom SSH job operation name.
	CustomSSHMode = "CustomSSH"
	// CustomAPIMode is custom Mikrotik secure API job operation name.
	CustomAPIMode = "CustomAPI"
	// InitSecureAPIMode is initialize secure API job operation name.
	InitSecureAPIMode = "InitSecureAPI"
	// InitPublicKeySSHMode is initialize public key SSH authentication job operation name.
	InitPublicKeySSHMode = "InitPublicKeySSH"
)
