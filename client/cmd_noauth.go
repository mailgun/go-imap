package client

import (
	"crypto/tls"
	"errors"
	"net"

	"github.com/emersion/go-sasl"
	"github.com/mailgun/go-imap"
	"github.com/mailgun/go-imap/commands"
	"github.com/mailgun/go-imap/responses"
)

var (
	// ErrAlreadyLoggedIn is returned if Login or Authenticate is called when the
	// client is already logged in.
	ErrAlreadyLoggedIn = errors.New("Already logged in")
	// ErrTLSAlreadyEnabled is returned if StartTLS is called when TLS is already
	// enabled.
	ErrTLSAlreadyEnabled = errors.New("TLS is already enabled")
	// ErrLoginDisabled is returned if Login or Authenticate is called when the
	// server has disabled authentication. Most of the time, calling enabling TLS
	// solves the problem.
	ErrLoginDisabled = errors.New("Login is disabled in current state")
)

// SupportStartTLS checks if the server supports STARTTLS.
func (c *Client) SupportStartTLS() (bool, error) {
	return c.Support(imap.StartTLS)
}

// StartTLS starts TLS negotiation.
func (c *Client) StartTLS(tlsConfig *tls.Config) error {
	if c.isTLS {
		return ErrTLSAlreadyEnabled
	}

	cmd := new(commands.StartTLS)

	err := c.Upgrade(func(conn net.Conn) (net.Conn, error) {
		if status, err := c.execute(cmd, nil); err != nil {
			return nil, err
		} else if err := status.Err(); err != nil {
			return nil, err
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			return nil, err
		}

		// Capabilities change when TLS is enabled
		c.capsLocker.Lock()
		c.caps = nil
		c.capsLocker.Unlock()

		return tlsConn, nil
	})
	if err != nil {
		return err
	}

	c.isTLS = true
	return nil
}

// SupportAuth checks if the server supports a given authentication mechanism.
func (c *Client) SupportAuth(mech string) (bool, error) {
	return c.Support("AUTH=" + mech)
}

// Authenticate indicates a SASL authentication mechanism to the server. If the
// server supports the requested authentication mechanism, it performs an
// authentication protocol exchange to authenticate and identify the client.
func (c *Client) Authenticate(auth sasl.Client) error {
	if c.State != imap.NotAuthenticatedState {
		return ErrAlreadyLoggedIn
	}

	mech, ir, err := auth.Start()
	if err != nil {
		return err
	}

	cmd := &commands.Authenticate{
		Mechanism: mech,
	}

	res := &responses.Authenticate{
		Mechanism:       auth,
		InitialResponse: ir,
		Writer:          c.Writer(),
	}

	status, err := c.execute(cmd, res)
	if err != nil {
		return err
	}
	if err = status.Err(); err != nil {
		return err
	}

	c.stateLocker.Lock()
	c.State = imap.AuthenticatedState
	c.stateLocker.Unlock()

	// Capabilities change when user is logged in
	c.capsLocker.Lock()
	c.caps = nil
	c.capsLocker.Unlock()

	if status.Code == imap.Capability {
		c.gotStatusCaps(status.Arguments)
	}

	return nil
}

// Login identifies the client to the server and carries the plaintext password
// authenticating this user.
func (c *Client) Login(username, password string) error {
	if c.State != imap.NotAuthenticatedState {
		return ErrAlreadyLoggedIn
	}

	c.capsLocker.Lock()
	loginDisabled := c.caps != nil && c.caps["LOGINDISABLED"]
	c.capsLocker.Unlock()
	if loginDisabled {
		return ErrLoginDisabled
	}

	cmd := &commands.Login{
		Username: username,
		Password: password,
	}

	status, err := c.execute(cmd, nil)
	if err != nil {
		return err
	}
	if err = status.Err(); err != nil {
		return err
	}

	c.stateLocker.Lock()
	c.State = imap.AuthenticatedState
	c.stateLocker.Unlock()

	c.capsLocker.Lock()
	c.caps = nil
	c.capsLocker.Unlock()

	if status.Code == "CAPABILITY" {
		c.gotStatusCaps(status.Arguments)
	}
	return nil
}
