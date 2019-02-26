package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const portSSH = "22"

// SSH defines SSH service by extending generic device with ssh specific client connection details.
type SSH struct {
	sshClient *ssh.Client
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
	sshconfig.Ciphers = append(sshconfig.Ciphers, "aes128-cbc", "3des-cbc")

	return &ssh.ClientConfig{
		Config:  sshconfig,
		Timeout: 30 * time.Second,
		User:    d.GetUser(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// HandleSequence establish connection and executes handler function with sequence of tasks/commands to do.
func (d *SSH) HandleSequence(handler HandlerFunc) (err error) {
	for idx, password := range d.GetPasswords() {
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
		break
	}

	if d.sshClient == nil {
		return fmt.Errorf("SSH no correct passwords found")
	}

	defer d.sshClient.Close()

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

// RunCmd runs a command using SSH and returns result.
func (d SSH) RunCmd(body string) (string, error) {
	session, err := d.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(body); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (d SSH) GetDevice() *GenericDevice {
	return &d.GenericDevice
}
