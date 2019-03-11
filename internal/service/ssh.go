package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const portSSH = "22"

// SSH defines SSH service by extending generic device with ssh specific client connection details.
type SSH struct {
	sshClient *ssh.Client
	session   *ssh.Session

	stdoutBuf io.Reader
	stdinBuf  io.WriteCloser
	prompt    *regexp.Regexp

	GenericDevice
}

func (d *SSH) GetUser() string {
	if d.Host.User != "" {
		return d.Host.User
	}
	return d.AppConfig.Service["ssh"].DefaultUser
}

func (d *SSH) GetPasswords() []string {
	list := func() string {
		if d.Host.Pass != "" {
			return d.Host.Pass
		}
		return d.AppConfig.Service["ssh"].DefaultPass
	}()

	return strings.Split(list, ",")
}

func (d *SSH) GetPort() string {
	if d.Host.Port != "" {
		return d.Host.Port
	}
	return portSSH
}

func (d *SSH) getConfig(password string) *ssh.ClientConfig {
	var sshconfig ssh.Config
	sshconfig.SetDefaults()
	sshconfig.Ciphers = append(sshconfig.Ciphers, "aes128-cbc", "aes192-cbc", "aes256-cbc", "3des-cbc", "des-cbc")

	return &ssh.ClientConfig{
		Config:  sshconfig,
		Timeout: 30 * time.Second,
		User:    d.GetUser() + "+ct", // Mikrotik hack to avoid detecting terminal capabilities and disable colors
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// HandleSequence establish connection and executes handler function with sequence of tasks/commands to do.
func (d *SSH) HandleSequence(ctx context.Context, handler HandlerFunc) (err error) {
	for idx, password := range d.GetPasswords() {
		select {
		case <-ctx.Done():
			return nil
		default:
			log.Printf("[IP:%s][SSH] Initializing connection :%s (using password #%d)", d.GenericDevice.Host.IP, d.GetPort(), idx)
			sshClientConfig := d.getConfig(password)

			// establish SSH connection
			d.sshClient, err = ssh.Dial("tcp", d.GenericDevice.Host.IP+":"+d.GetPort(), sshClientConfig)
			if err != nil && err.Error() == "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none password], no supported methods remain" {
				continue
			}
			if err != nil {
				return fmt.Errorf("SSH handle sequence error %v", err)
			}
		}
		// store valid password for this device
		d.GenericDevice.Host.Pass = password
		break
	}

	if d.sshClient == nil {
		return fmt.Errorf("SSH no correct passwords found")
	}

	defer d.sshClient.Close()
	defer d.Close()

	d.matches = make(map[string]string)

	// call handlr with sequence of operations
	return handler(d)
}

// CopyFile copies file over SFTP.
func (d *SSH) CopyFile(local, remote string) error {
	client, err := sftp.NewClient(d.sshClient)
	if err != nil {
		return fmt.Errorf("SFTP client creating error error %v", err)
	}
	defer client.Close()

	// file on remote device
	rf, err := client.OpenFile(remote, os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("SFTP remote file %s error %v", remote, err)
	}
	defer rf.Close()

	// local file
	lf, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("Local file %s open error %v", local, err)
	}
	defer lf.Close()

	io.Copy(rf, lf)

	log.Printf("[IP:%s][SFTP] %s copied to %s\n", d.GenericDevice.Host.IP, local, remote)
	return nil
}

// Close used session
func (d *SSH) Close() error {
	defer d.session.Close()
	d.stdinBuf.Write([]byte("/quit\r"))

	if err := d.session.Wait(); err != nil {
		return fmt.Errorf("remote command did not exit cleanly: %v", err)
	}
	return nil
}

// RunCmd runs a command using SSH and returns result.
func (d *SSH) RunCmd(body string, expect *regexp.Regexp) (result string, err error) {
	if err = d.initializeSession(); err != nil {
		return
	}

	d.stdinBuf.Write([]byte(body + "\r"))

	if expect != nil {
		result, err = d.expect(d.stdoutBuf, expect)
	} else {
		result, err = d.expect(d.stdoutBuf, d.prompt)
	}

	return
}

func (d *SSH) GetDevice() *GenericDevice {
	return &d.GenericDevice
}

func (d *SSH) initializeSession() (err error) {
	if d.session != nil {
		return nil
	}

	d.prompt = regexp.MustCompile(`(?sm)\[.*@.*\] >.{0,1}$`)

	if d.session, err = d.sshClient.NewSession(); err != nil {
		return fmt.Errorf("session initializing error: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}
	d.session.Stderr = os.Stderr

	if err = d.session.RequestPty("xterm", 500, 200, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}
	if d.stdoutBuf, err = d.session.StdoutPipe(); err != nil {
		return fmt.Errorf("request for stdout pipe failed: %s", err)
	}
	if d.stdinBuf, err = d.session.StdinPipe(); err != nil {
		return fmt.Errorf("request for stdin pipe failed: %s", err)
	}
	if err = d.session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %s", err)
	}

	_, err = d.expect(d.stdoutBuf, d.prompt)
	return
}
