package main

import (
	"errors"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	oauthdialog "github.com/emersion/go-oauthdialog"
	"github.com/emersion/go-sasl"
	"golang.org/x/oauth2"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.mail.yahoo.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("mixas8383@yahoo.com", "btlcatkgunyindfr"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
}
func authenticate(c *client.Client, cfg *oauth2.Config, username string) error {
	ok, _ := c.SupportAuth(sasl.Xoauth2)
	if ok {
		return errors.New("XOAUTH2 not supported by the server")
	}

	// Ask for the user to login with his Google account
	code, err := oauthdialog.Open(cfg)
	if err != nil {
		return err
	}

	// Get a token from the returned code
	// This token can be saved in a secure store to be reused later
	token, err := cfg.Exchange(oauth2.NoContext, code)
	if err != nil {
		return err
	}

	// Login to the IMAP server with XOAUTH2
	saslClient := sasl.NewXoauth2Client(username, token.AccessToken)
	return c.Authenticate(saslClient)
}
