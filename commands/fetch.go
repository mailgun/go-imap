package commands

import (
	"errors"
	"strings"

	"github.com/mailgun/go-imap"
)

// Fetch is a FETCH command, as defined in RFC 3501 section 6.4.5.
type Fetch struct {
	SeqSet *imap.SeqSet
	Items  []string
}

func (cmd *Fetch) Command() *imap.Command {
	items := make([]interface{}, len(cmd.Items))
	for i, item := range cmd.Items {
		if section, err := imap.NewBodySectionName(item); err == nil {
			items[i] = section
		} else {
			items[i] = item
		}
	}

	return &imap.Command{
		Name:      imap.Fetch,
		Arguments: []interface{}{cmd.SeqSet, items},
	}
}

func (cmd *Fetch) Parse(fields []interface{}) error {
	if len(fields) < 2 {
		return errors.New("No enough arguments")
	}

	seqset, ok := fields[0].(string)
	if !ok {
		return errors.New("Sequence set must be a string")
	}

	var err error
	if cmd.SeqSet, err = imap.NewSeqSet(seqset); err != nil {
		return err
	}

	switch items := fields[1].(type) {
	case string: // A macro or a single item
		switch strings.ToUpper(items) {
		case "ALL":
			cmd.Items = []string{"FLAGS", "INTERNALDATE", "RFC822.SIZE", "ENVELOPE"}
		case "FAST":
			cmd.Items = []string{"FLAGS", "INTERNALDATE", "RFC822.SIZE"}
		case "FULL":
			cmd.Items = []string{"FLAGS", "INTERNALDATE", "RFC822.SIZE", "ENVELOPE", "BODY"}
		default:
			cmd.Items = []string{strings.ToUpper(items)}
		}
	case []interface{}: // A list of items
		cmd.Items = make([]string, len(items))
		for i, v := range items {
			item, _ := v.(string)
			cmd.Items[i] = strings.ToUpper(item)
		}
	default:
		return errors.New("Items must be either a string or a list")
	}

	return nil
}
