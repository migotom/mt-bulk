package mode

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/vulnerabilities"
)

// SecurityAudit is an operation performing security audit of device.
func SecurityAudit(vulnerabilitiesManager *vulnerabilities.Manager) OperationModeFunc {
	return func(ctx context.Context, sugar *zap.SugaredLogger, client clients.Client, job *entities.Job) entities.Result {
		results := make([]entities.CommandResult, 0, 9)

		establishResult, err := clients.EstablishConnection(ctx, sugar, client, job)
		results = append(results, establishResult)
		if err != nil {
			return entities.Result{Errors: []error{err}}
		}
		defer client.Close()

		// prepare sequence of commands to run on device
		commands := []entities.Command{
			{Body: `/system resource print`, MatchPrefix: "v", Match: `(?m)\s+version:\s+([\d\.]+)+`},
			{Body: `/ip service print`, MatchPrefix: "service",
				Matches: []string{`(?m)\d+\s+(telnet)`, `(?m)\d+\s+(ftp)`, `(?m)\d+\s+(www)`, `(?m)\d+\s+(ftp)`, `(?m)\d+\s+(www)`, `(?m)\d+\s+(api)`, `(?m)\d+\s+(api-ssl)`}},
			{Body: `/tool mac-server print`, MatchPrefix: "mac-server", Match: `(?m)\s+(allowed-interface-list:\s+[^n][^o][^n][^e])`},
			{Body: `/tool mac-server mac-winbox print`, MatchPrefix: "mac-winbox", Match: `(?m)\s+(allowed-interface-list:\s+[^n][^o][^n][^e])`},
			{Body: `/tool mac-server ping print`, MatchPrefix: "mac-ping", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/ip neighbor discovery-settings print`, MatchPrefix: "neighbor", Match: `(?m)\s+(discover-interface-list:\s+[^n][^o][^n][^e])`},
			{Body: `/tool bandwidth-server print`, MatchPrefix: "bandwidth-server", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/ip dns print`, MatchPrefix: "dns", Match: `(?m)\s+(allow-remote-requests:\s+yes)`},
			{Body: `/ip proxy print`, MatchPrefix: "proxy", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/ip socks print`, MatchPrefix: "socks", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/ip upnp print`, MatchPrefix: "socks", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/tool romon print`, MatchPrefix: "romon", Match: `(?m)\s+(enabled:\s+yes)`},
			{Body: `/user print`, MatchPrefix: "admin-full", Match: `(?m)\s+(\d+\s+(?:;;; system default user)?\s+admin\s+full)`},
			{Body: `/snmp community print where name=public`, MatchPrefix: "snmp", Match: `(?m)\s+(public\s+::\/0)|(public\s+([\d\.\/:]+)\s+none)`},
			{Body: `/ip ssh print`, MatchPrefix: "ssh", Match: `(?m)\s+(strong-crypto:\s+no)`},
			{Body: `/ip settings print`, MatchPrefix: "rp-filter", Match: `(?m)\s+(rp-filter:\s+no)`},
		}

		commandResults, matches, err := clients.ExecuteCommands(ctx, client, commands)
		results = append(results, commandResults...)
		if err != nil {
			return entities.Result{Results: results, Errors: []error{fmt.Errorf("executing SecurityAudit commands error %v", err)}}
		}

		var optionsErr optionsError

		unsecureServices := extractUnsecureOptions(regexp.MustCompile(`%{service\d+}`), matches)
		if len(unsecureServices) > 0 {
			optionsErr.err = append(optionsErr.err, fmt.Errorf("enabled services %v", unsecureServices))
		}

		singleTests := []struct {
			re  string
			err string
		}{
			{`%{mac-server\d+}`, "enabled mac-server"},
			{`%{mac-ping\d+}`, "enabled ping by mac"},
			{`%{neighbor\d+}`, "enabled neighbor discovery"},
			{`%{bandwidth-server\d+}`, "enabled bandwidth server"},
			{`%{dns\d+}`, "DNS server allows remote requests"},
			{`%{proxy\d+}`, "enabled proxy server"},
			{`%{socks\d+}`, "enabled socks server"},
			{`%{upnp\d+}`, "enabled upnp server"},
			{`%{romon\d+}`, "enabled RoMON agent"},
			{`%{rp-filter\d+}`, "Reverse Path Filtering not enabled"},
			{`%{snmp\d+}`, "SNMP publicly available"},
			{`%{ssh\d+}`, "not enabled SSH strong-crypto"},
			{`%{admin-full\d+}`, "enabled admin user without allowed IP restriction with full grants"},
		}

		for _, test := range singleTests {
			if extractUnsecureOption(regexp.MustCompile(test.re), matches) != "" {
				optionsErr.err = append(optionsErr.err, errors.New(test.err))
			}
		}

		version, ok := matches["(%{v1})"]
		if !ok {
			return entities.Result{Results: results, Errors: []error{errors.New("Mikrotik version not recognized"), optionsErr}}
		}
		vulnerabilitiesErrors := vulnerabilitiesManager.Check(version)

		return entities.Result{Results: results, Errors: []error{optionsErr, vulnerabilitiesErrors}}
	}
}

func extractUnsecureOptions(re *regexp.Regexp, matches map[string]string) (options []string) {
	for match, value := range matches {
		if re.MatchString(match) {
			options = append(options, value)
		}
	}
	return
}

func extractUnsecureOption(re *regexp.Regexp, matches map[string]string) (option string) {
	for match, value := range matches {
		if re.MatchString(match) {
			return value
		}
	}
	return
}

type SecurityAuditError struct {
	OptionsErrors         error
	VulnerabilitiesErrors error
}

func (ser SecurityAuditError) Error() string {
	var err strings.Builder
	if ser.OptionsErrors != nil {
		err.WriteString(fmt.Sprintf("%v\t", ser.OptionsErrors))
	}
	if ser.VulnerabilitiesErrors != nil {
		err.WriteString(ser.VulnerabilitiesErrors.Error())
	}
	return err.String()
}

type optionsError struct {
	err []error
}

func (ser optionsError) Error() string {
	var err strings.Builder
	err.WriteString("unsecure options found: ")

	for _, auditErr := range ser.err {
		err.WriteString(fmt.Sprintf("%s, ", auditErr))
	}

	return strings.TrimRight(err.String(), ", ")
}
