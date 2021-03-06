package commands

import (
	"errors"

	"github.com/mailgun/go-imap"
	"github.com/mailgun/go-imap/utf7"
)

// Delete is a DELETE command, as defined in RFC 3501 section 6.3.3.
type Delete struct {
	Mailbox string
}

func (cmd *Delete) Command() *imap.Command {
	mailbox, _ := utf7.Encoder.String(cmd.Mailbox)

	return &imap.Command{
		Name:      imap.Delete,
		Arguments: []interface{}{mailbox},
	}
}

func (cmd *Delete) Parse(fields []interface{}) error {
	if len(fields) < 1 {
		return errors.New("No enough arguments")
	}

	if mailbox, ok := fields[0].(string); !ok {
		return errors.New("Mailbox name must be a string")
	} else if mailbox, err := utf7.Decoder.String(mailbox); err != nil {
		return err
	} else {
		cmd.Mailbox = imap.CanonicalMailboxName(mailbox)
	}

	return nil
}
