package emailclient

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

const (
	defaultDialTimeout = 30 * time.Second
	// defaultHandshakeTimeout bounds the implicit-TLS handshake, the one exchange
	// imapclient does not time itself. Everything after it, imapclient bounds on
	// its own: 30s for a command write or a response, 5 minutes for a literal.
	defaultHandshakeTimeout = 30 * time.Second
	// headerFetchBatchSize keeps each UID FETCH command line short, since a
	// scattered UID set does not compress into ranges.
	headerFetchBatchSize = 200
)

// Credentials identify one IMAP account.
type Credentials struct {
	Host        string
	Port        string
	Username    string
	Password    string
	UseStartTLS bool
}

// MessageHeader is one unread message's routing fields, fetched without its body.
type MessageHeader struct {
	UID uint32
	// Size is the full message size in bytes as reported by the server.
	Size int64
	// Header is the raw From and Subject header block.
	Header []byte
}

// Mailbox is the IMAP surface the poller needs. Nothing it does marks a message
// read except MarkSeen.
type Mailbox interface {
	// UnseenUIDs lists the INBOX messages without the \Seen flag, in UID order.
	UnseenUIDs() ([]uint32, error)
	// FetchHeaders returns the routing header and size of each message.
	FetchHeaders(uids []uint32) ([]MessageHeader, error)
	// FetchRaw returns the whole message.
	FetchRaw(uid uint32) ([]byte, error)
	// MarkSeen sets \Seen on the messages.
	MarkSeen(uids ...uint32) error
	// MarkUnseen clears \Seen, handing the messages back to the next poll.
	MarkUnseen(uids ...uint32) error
	Close() error
}

// MessageFetchError is a problem with one message rather than with the session,
// which stays usable: the server no longer has the message, or would not return
// it. Any other error from FetchRaw means the session itself failed.
type MessageFetchError struct {
	UID uint32
	// Gone means the server returned nothing for the message, typically because
	// another client expunged it after it was listed.
	Gone bool
	// Temporary means the server asked to be asked again later (UNAVAILABLE).
	Temporary bool
	Err       error
}

func (err *MessageFetchError) Error() string {
	return fmt.Sprintf("message %d: %s", err.UID, err.Err.Error())
}

func (err *MessageFetchError) Unwrap() error {
	return err.Err
}

// Connector opens IMAP sessions. The zero value is ready for production use.
type Connector struct {
	// TLSConfig is the base TLS configuration. Nil means the system roots with
	// full certificate verification. ServerName defaults to the host.
	TLSConfig *tls.Config
	// Dial opens the TCP connection. Nil means a net.Dialer with a 30s timeout.
	Dial func(network string, address string) (net.Conn, error)
	// HandshakeTimeout overrides the limit on the implicit-TLS handshake.
	HandshakeTimeout time.Duration
}

// Connect logs in and selects INBOX. Like the Python client it uses implicit TLS
// unless the account asks for STARTTLS, and always verifies the certificate.
func (connector Connector) Connect(credentials Credentials) (Mailbox, error) {
	address := net.JoinHostPort(credentials.Host, credentials.Port)

	rawConn, err := connector.dial("tcp", address)
	if err != nil {
		return nil, err
	}

	tlsConfig := connector.tlsConfig(credentials.Host)
	options := &imapclient.Options{TLSConfig: tlsConfig}

	var client *imapclient.Client
	if credentials.UseStartTLS {
		client, err = imapclient.NewStartTLS(rawConn, options)
		if err != nil {
			rawConn.Close()
			return nil, err
		}
	} else {
		tlsConn, err := connector.handshake(rawConn, tlsConfig)
		if err != nil {
			rawConn.Close()
			return nil, err
		}
		client = imapclient.New(tlsConn, options)
	}

	if err := client.Login(credentials.Username, credentials.Password).Wait(); err != nil {
		client.Close()
		return nil, err
	}

	if _, err := client.Select("INBOX", nil).Wait(); err != nil {
		client.Close()
		return nil, err
	}

	return &imapMailbox{client: client}, nil
}

// CheckConnection connects, logs in, selects INBOX and logs out, returning the
// first error.
func (connector Connector) CheckConnection(credentials Credentials) error {
	mailbox, err := connector.Connect(credentials)
	if err != nil {
		return err
	}
	return mailbox.Close()
}

func (connector Connector) dial(network string, address string) (net.Conn, error) {
	if connector.Dial != nil {
		return connector.Dial(network, address)
	}
	dialer := &net.Dialer{Timeout: defaultDialTimeout}
	return dialer.Dial(network, address)
}

// handshake runs the implicit-TLS handshake under a deadline, then clears it.
// From there on imapclient sets its own deadline before every read and write,
// and waits for the next response with none at all; nothing here may push one
// onto the connection, or the session dies whenever the poller spends that long
// between two commands. Clearing this one keeps that true without relying on
// imapclient to overwrite it.
func (connector Connector) handshake(conn net.Conn, config *tls.Config) (*tls.Conn, error) {
	if err := conn.SetDeadline(time.Now().Add(connector.handshakeTimeout())); err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, config)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}

	return tlsConn, nil
}

func (connector Connector) handshakeTimeout() time.Duration {
	if connector.HandshakeTimeout > 0 {
		return connector.HandshakeTimeout
	}
	return defaultHandshakeTimeout
}

func (connector Connector) tlsConfig(host string) *tls.Config {
	var config *tls.Config
	if connector.TLSConfig != nil {
		config = connector.TLSConfig.Clone()
	} else {
		config = &tls.Config{}
	}
	if len(config.ServerName) == 0 {
		config.ServerName = host
	}
	return config
}

type imapMailbox struct {
	client *imapclient.Client
}

func (mailbox *imapMailbox) UnseenUIDs() ([]uint32, error) {
	data, err := mailbox.client.UIDSearch(&imap.SearchCriteria{NotFlag: []imap.Flag{imap.FlagSeen}}, nil).Wait()
	if err != nil {
		return nil, err
	}

	uids := make([]uint32, 0)
	for _, uid := range data.AllUIDs() {
		uids = append(uids, uint32(uid))
	}
	slices.Sort(uids)

	return uids, nil
}

func (mailbox *imapMailbox) FetchHeaders(uids []uint32) ([]MessageHeader, error) {
	section := &imap.FetchItemBodySection{
		Specifier:    imap.PartSpecifierHeader,
		HeaderFields: []string{"From", "Subject"},
		Peek:         true,
	}

	headers := make([]MessageHeader, 0, len(uids))
	for start := 0; start < len(uids); start += headerFetchBatchSize {
		end := min(start+headerFetchBatchSize, len(uids))

		messages, err := mailbox.client.Fetch(toUIDSet(uids[start:end]), &imap.FetchOptions{
			UID:         true,
			RFC822Size:  true,
			BodySection: []*imap.FetchItemBodySection{section},
		}).Collect()
		if err != nil {
			return nil, err
		}

		for _, message := range messages {
			headers = append(headers, MessageHeader{
				UID:    uint32(message.UID),
				Size:   message.RFC822Size,
				Header: message.FindBodySection(section),
			})
		}
	}

	slices.SortFunc(headers, func(a, b MessageHeader) int { return int(a.UID) - int(b.UID) })
	return headers, nil
}

func (mailbox *imapMailbox) FetchRaw(uid uint32) ([]byte, error) {
	// BODY.PEEK[] rather than RFC822, which sets \Seen as a side effect: a message
	// only becomes read once it has been handled.
	section := &imap.FetchItemBodySection{Peek: true}

	messages, err := mailbox.client.Fetch(toUIDSet([]uint32{uid}), &imap.FetchOptions{
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{section},
	}).Collect()
	if err != nil {
		// A NO is the server declining this one message. A BAD is a malformed
		// command, which would fail for every message alike, so like a broken
		// connection it is the session's problem.
		var imapErr *imap.Error
		if errors.As(err, &imapErr) && imapErr.Type == imap.StatusResponseTypeNo {
			return nil, &MessageFetchError{UID: uid, Temporary: imapErr.Code == imap.ResponseCodeUnavailable, Err: err}
		}
		return nil, err
	}

	for _, message := range messages {
		if uint32(message.UID) == uid {
			if body := message.FindBodySection(section); body != nil {
				return body, nil
			}
		}
	}

	return nil, &MessageFetchError{UID: uid, Gone: true, Err: errors.New("the server did not return it")}
}

func (mailbox *imapMailbox) MarkSeen(uids ...uint32) error {
	return mailbox.storeSeen(imap.StoreFlagsAdd, uids)
}

func (mailbox *imapMailbox) MarkUnseen(uids ...uint32) error {
	return mailbox.storeSeen(imap.StoreFlagsDel, uids)
}

func (mailbox *imapMailbox) storeSeen(op imap.StoreFlagsOp, uids []uint32) error {
	if len(uids) == 0 {
		return nil
	}

	return mailbox.client.Store(toUIDSet(uids), &imap.StoreFlags{
		Op:     op,
		Silent: true,
		Flags:  []imap.Flag{imap.FlagSeen},
	}, nil).Close()
}

func (mailbox *imapMailbox) Close() error {
	logoutErr := mailbox.client.Logout().Wait()
	closeErr := mailbox.client.Close()
	return errors.Join(logoutErr, closeErr)
}

func toUIDSet(uids []uint32) imap.UIDSet {
	converted := make([]imap.UID, len(uids))
	for i, uid := range uids {
		converted[i] = imap.UID(uid)
	}
	return imap.UIDSetNum(converted...)
}
