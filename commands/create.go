package commands

import (
	"errors"

	"github.com/mailgun/go-imap"
	"github.com/mailgun/go-imap/utf7"
)

// Create is a CREATE command, as defined in RFC 3501 section 6.3.3.
type Create struct {
	Mailbox string
}

func (cmd *Create) Command() *imap.Command {
	mailbox, _ := utf7.Encoder.String(cmd.Mailbox)

	return &imap.Command{
		Name:      imap.Create,
		Arguments: []interface{}{mailbox},
	}
}

func (cmd *Create) Parse(fields []interface{}) error {
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
