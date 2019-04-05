// A memory backend.
package memory

import (
	"errors"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
)

type Backend struct {
	users map[string]*User
}

func (be *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	user, ok := be.users[username]
	if ok && user.password == password {
		return user, nil
	}

	return nil, errors.New("Bad username or password")
}

func New() *Backend {
	user := &User{username: "username", password: "password"}

	body := "From: contact@example.org\r\n" +
		"To: contact@example.org\r\n" +
		"Subject: A little message, just for you\r\n" +
		"Date: Wed, 11 May 2016 14:31:59 +0000\r\n" +
		"Message-ID: <0000000@localhost/>\r\n" +
		"Content-Type: text/plain\r\n" +
		"\r\n" +
		"Hi there :)"

	user.mailboxes = map[string]*Mailbox{
		"INBOX": {
			name: "INBOX",
			user: user,
			Messages: []*Message{
				{
					Uid:   6,
					Date:  time.Now(),
					Flags: []string{"\\Seen"},
					Size:  uint32(len(body)),
					Body:  []byte(body),
				},
			},
		},
	}

	return &Backend{
		users: map[string]*User{user.username: user},
	}
}

func (be *Backend) Push(mime string, t time.Time, uid uint32) {
	msg := &Message{
		Uid:   uid,
		Date:  t,
		Flags: []string{"\\Seen"},
		Size:  uint32(len(mime)),
		Body:  []byte(mime),
	}
	be.users["username"].mailboxes["INBOX"].Messages = append(be.users["username"].mailboxes["INBOX"].Messages, msg)
}

func (be *Backend) Pop() {
	count := len(be.users["username"].mailboxes["INBOX"].Messages)
	be.users["username"].mailboxes["INBOX"].Messages = be.users["username"].mailboxes["INBOX"].Messages[:count-1]
}
