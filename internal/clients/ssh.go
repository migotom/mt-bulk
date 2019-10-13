package clients

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pkg/sftp"
	cryptossh "golang.org/x/crypto/ssh"
)

// SSHDefaultPort is default SSH server port.
const SSHDefaultPort = "22"

// NewSSHClient returns new SSH client.
func NewSSHClient(config Config) Client {
	return &SSH{
		prompt: regexp.MustCompile(`(?sm)\x1b?\[.*@.*\] >.{0,1}$`),
		Config: config,
	}
}

// SSH defines SSH client.
type SSH struct {
	client  *cryptossh.Client
	session *cryptossh.Session

	stdoutBuf io.Reader
	stdinBuf  io.Writer
	prompt    *regexp.Regexp

	Config
}

// GetConfig returns client's configuration.
func (ssh *SSH) GetConfig() Config {
	return ssh.Config
}

// Connect to device using SSH.
func (ssh *SSH) Connect(ctx context.Context, IP, Port, User, Password string) (err error) {
	var sshConfig cryptossh.Config
	var sshAuthMethods []cryptossh.AuthMethod

	if key, err := ioutil.ReadFile(filepath.Join(ssh.Config.KeyStore, "id_rsa.key")); err == nil {
		if signer, err := cryptossh.ParsePrivateKey(key); err == nil {
			sshAuthMethods = append(sshAuthMethods, cryptossh.PublicKeys(signer))
		}
	}
	sshAuthMethods = append(sshAuthMethods, cryptossh.Password(Password))

	sshConfig.SetDefaults()
	sshConfig.Ciphers = append(sshConfig.Ciphers, "aes128-cbc", "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes192-cbc", "aes256-cbc", "3des-cbc", "des-cbc", "diffie-hellman-group-exchange-sha256")
	sshConfig.KeyExchanges = append(sshConfig.KeyExchanges, "diffie-hellman-group-exchange-sha256")
	clientConfig := &cryptossh.ClientConfig{
		Config:          sshConfig,
		Timeout:         30 * time.Second,
		User:            User + "+ct", // Mikrotik hack to avoid detecting terminal capabilities and disable colors
		Auth:            sshAuthMethods,
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
	}

	ssh.client, err = cryptossh.Dial("tcp", fmt.Sprintf("%s:%s", IP, Port), clientConfig)
	if err != nil && err.Error() == "ssh: handshake failed: ssh: unable to authenticate, attempted methods [none password], no supported methods remain" {
		return ErrorWrongPassword{err}
	}
	if err != nil {
		return ErrorRetryable{fmt.Errorf("SSH handle sequence error %v", err)}
	}
	return nil
}

// CopyFile copies file over SFTP.
func (ssh *SSH) CopyFile(ctx context.Context, local, remote string) (result string, err error) {
	errorChan := make(chan error)
	responseChan := make(chan string)

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		sftpClient, err := sftp.NewClient(ssh.client)
		if err != nil {
			errorChan <- fmt.Errorf("SFTP client creating error error %v", err)
		}
		defer sftpClient.Close()

		// file on remote device
		rf, err := sftpClient.OpenFile(remote, os.O_CREATE|os.O_WRONLY)
		if err != nil {
			errorChan <- fmt.Errorf("SFTP remote file %s error %v", remote, err)
		}
		defer rf.Close()

		// local file
		lf, err := os.Open(local)
		if err != nil {
			errorChan <- fmt.Errorf("local file %s open error %v", local, err)
		}
		defer lf.Close()

		io.Copy(rf, lf)

		responseChan <- fmt.Sprintf("/<mt-bulk>copy sftp://%s %s", local, remote)
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context cancelled")
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("copy file timeouted")
	case err := <-errorChan:
		return "", err
	case response := <-responseChan:
		return response, err
	}
}

// Close SSH client session.
func (ssh *SSH) Close() error {
	if ssh.client == nil {
		return nil
	}
	defer ssh.client.Close()
	defer func() {
		ssh.session.Close()
		ssh.session = nil
	}()

	ssh.stdinBuf.Write([]byte("/quit\r"))
	if err := ssh.session.Wait(); err != nil {
		return fmt.Errorf("remote command did not exit cleanly: %v", err)
	}
	return nil
}

// RunCmd execues given command on remote device, optionally can compare execution result with provided expect regexp.
func (ssh *SSH) RunCmd(body string, expect *regexp.Regexp) (result string, err error) {
	if err = ssh.initializeSession(); err != nil {
		return
	}

	ssh.stdinBuf.Write([]byte(body + "\r"))

	if expect != nil {
		result, err = waitForExpected(ssh.stdoutBuf, expect)
	} else {
		result, err = waitForExpected(ssh.stdoutBuf, ssh.prompt)
	}
	return ssh.prompt.ReplaceAllString(result, ""), err
}

func (ssh *SSH) initializeSession() (err error) {
	if ssh.session != nil {
		return nil
	}

	if ssh.session, err = ssh.client.NewSession(); err != nil {
		return fmt.Errorf("session initializing error: %s", err)
	}

	modes := cryptossh.TerminalModes{
		cryptossh.ECHO:          0,
		cryptossh.ECHOCTL:       0,
		cryptossh.TTY_OP_ISPEED: 115200,
		cryptossh.TTY_OP_OSPEED: 115200,
	}
	ssh.session.Stderr = os.Stderr

	if err = ssh.session.RequestPty("xterm", ssh.Pty.Height, ssh.Pty.Widht, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}
	if ssh.stdinBuf, err = ssh.session.StdinPipe(); err != nil {
		return fmt.Errorf("request for stdin pipe failed: %s", err)
	}
	if ssh.stdoutBuf, err = ssh.session.StdoutPipe(); err != nil {
		return fmt.Errorf("request for stdout pipe failed: %s", err)
	}

	if err = ssh.session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %s", err)
	}

	_, err = waitForExpected(ssh.stdoutBuf, ssh.prompt)
	return
}
