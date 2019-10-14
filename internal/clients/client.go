package clients

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/entities"
)

// Client interface for all supported clients.
type Client interface {
	GetConfig() Config

	RunCmd(string, *regexp.Regexp) (string, error)
	Connect(ctx context.Context, IP, Port, User, Password string) error
	Close() error
}

// Copier interface for copy files capable clients.
type Copier interface {
	CopyFile(ctx context.Context, local, remote string) (string, error)
}

// EstablishConnection tries to establish connection for provided host by specified client.
// It tries to connect by retries number and list of passwords defined in client's configuration.
func EstablishConnection(ctx context.Context, sugar *zap.SugaredLogger, client Client, job *entities.Job) (err error) {
	err = errors.New("unexpected error during connection establishing")

	defer func() {
		if err != nil {
			err = fmt.Errorf("could not be able to establish connection (%s) (%v)", job.Host, err)
		}
	}()

	config := client.GetConfig()
	job.Host.SetDefaults(config.DefaultPort, config.DefaultUser, config.DefaultPassword)
	// TODO add verification that remote device is mt-bulk comaptible (eg. Mikrotik SSH, Mikrotik API)
	for retry := 0; retry < config.Retries; retry++ {
		time.Sleep(time.Duration(retry*retry) * time.Millisecond * 100)

		for idx, password := range job.Host.GetPasswords() {

			select {
			case <-ctx.Done():
				return errors.New("context done")

			default:
				sugar.Infow("establishing connection", "host", job.Host, "password id", idx, "retry", retry, "id", job.ID)

				err = client.Connect(ctx, job.Host.IP, job.Host.Port, job.Host.User, password)
				if err != nil {
					if _, ok := err.(ErrorWrongPassword); ok {
						continue
					}
					if _, ok := err.(ErrorRetryable); ok {
						break
					}
				}
				return
			}

			// store valid password for this device
			job.Host.Password = password
		}
	}
	return
}

// ExecuteCommands executes provided list of commands using specified client.
func ExecuteCommands(ctx context.Context, d Client, commands []entities.Command) ([]entities.CommandResult, error) {
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
			responseChan <- result
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

	var executeError error
	executed := make([]entities.CommandResult, 0, len(commands))

	for _, c := range commands {
		errChan := make(chan error)
		responseChan := make(chan string)

		go run(c, responseChan, errChan)

		commandResult := entities.CommandResult{Body: c.Body}
	commandParseLoop:
		for {
			select {
			case <-ctx.Done():
				return executed, errors.New("interrupted")
			case <-time.After(30 * time.Second):
				return executed, fmt.Errorf("timeouted")
			case err, ok := <-errChan:
				if ok {
					executeError = fmt.Errorf("%v (%s)", err, c)
				}
			case response, ok := <-responseChan:
				if !ok {
					break commandParseLoop
				}
				commandResult.Responses = append(commandResult.Responses, response)
			}
		}
		commandResult.Error = executeError
		executed = append(executed, commandResult)

		if executeError != nil {
			break
		}
	}
	return executed, executeError
}
