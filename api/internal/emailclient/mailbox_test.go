package emailclient

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"
)

const (
	testUsername = "receipts@example.org"
	testPassword = "correct horse"
)

type testServer struct {
	host    string
	port    string
	user    *imapmemserver.User
	rootCAs *x509.CertPool
}

// newSelfSignedCert issues a certificate for 127.0.0.1 and the pool trusting it.
func newSelfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "imap.test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(certificate)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, pool
}

// startServer serves an in-memory mailbox over implicit TLS, or over plain TCP
// offering STARTTLS.
func startServer(t *testing.T, startTLS bool) testServer {
	t.Helper()
	return startServerWith(t, startTLS, nil)
}

// startServerWith is startServer with each session passed through wrap, so a
// test can make the server misbehave.
func startServerWith(t *testing.T, startTLS bool, wrap func(imapserver.Session) imapserver.Session) testServer {
	t.Helper()
	certificate, pool := newSelfSignedCert(t)
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{certificate}}

	memServer := imapmemserver.New()
	user := imapmemserver.NewUser(testUsername, testPassword)
	if err := user.Create("INBOX", nil); err != nil {
		t.Fatal(err)
	}
	memServer.AddUser(user)

	options := &imapserver.Options{
		NewSession: func(conn *imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			session := memServer.NewSession()
			if wrap != nil {
				session = wrap(session)
			}
			return session, nil, nil
		},
		InsecureAuth: true,
		Caps:         imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapIMAP4rev2: {}},
	}

	var listener net.Listener
	var err error
	if startTLS {
		options.TLSConfig = tlsConfig
		listener, err = net.Listen("tcp", "127.0.0.1:0")
	} else {
		listener, err = tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	}
	if err != nil {
		t.Fatal(err)
	}

	server := imapserver.New(options)
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })

	host, port, _ := net.SplitHostPort(listener.Addr().String())
	return testServer{host: host, port: port, user: user, rootCAs: pool}
}

func (server testServer) credentials(startTLS bool) Credentials {
	return Credentials{Host: server.host, Port: server.port, Username: testUsername, Password: testPassword, UseStartTLS: startTLS}
}

func (server testServer) connector() Connector {
	return Connector{TLSConfig: &tls.Config{RootCAs: server.rootCAs}}
}

func (server testServer) deliver(t *testing.T, raw []byte, flags ...imap.Flag) {
	t.Helper()
	if _, err := server.user.Append("INBOX", bytes.NewReader(raw), &imap.AppendOptions{Flags: flags}); err != nil {
		t.Fatal(err)
	}
}

func connect(t *testing.T, server testServer, startTLS bool) Mailbox {
	t.Helper()
	mailbox, err := server.connector().Connect(server.credentials(startTLS))
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	t.Cleanup(func() { mailbox.Close() })
	return mailbox
}

func TestMailbox_UnreadMessagesStayUnreadUntilMarked(t *testing.T) {
	for _, startTLS := range []bool{false, true} {
		name := "implicit tls"
		if startTLS {
			name = "starttls"
		}

		t.Run(name, func(t *testing.T) {
			server := startServer(t, startTLS)
			first := receiptFrom("bob@example.com")
			server.deliver(t, first)
			server.deliver(t, receiptFrom("old@example.com"), imap.FlagSeen)
			server.deliver(t, receiptFrom("alice@example.com"))

			mailbox := connect(t, server, startTLS)

			unseen, err := mailbox.UnseenUIDs()
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(unseen, []uint32{1, 3}) {
				t.Fatalf("UnseenUIDs() = %v, want [1 3]", unseen)
			}

			raw, err := mailbox.FetchRaw(1)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, first) {
				t.Errorf("FetchRaw() returned %d bytes that differ from the delivered message", len(raw))
			}

			// Fetching must not mark the message read: the Python client's
			// RFC822 fetch did, which is how a failed poll lost its batch.
			unseen, _ = mailbox.UnseenUIDs()
			if !slices.Equal(unseen, []uint32{1, 3}) {
				t.Fatalf("after FetchRaw, UnseenUIDs() = %v, want [1 3]", unseen)
			}

			if err := mailbox.MarkSeen(1); err != nil {
				t.Fatal(err)
			}
			unseen, _ = mailbox.UnseenUIDs()
			if !slices.Equal(unseen, []uint32{3}) {
				t.Errorf("after MarkSeen(1), UnseenUIDs() = %v, want [3]", unseen)
			}

			if err := mailbox.MarkUnseen(1); err != nil {
				t.Fatal(err)
			}
			unseen, _ = mailbox.UnseenUIDs()
			if !slices.Equal(unseen, []uint32{1, 3}) {
				t.Errorf("after MarkUnseen(1), UnseenUIDs() = %v, want [1 3]", unseen)
			}
		})
	}
}

func TestMailbox_FetchHeaders(t *testing.T) {
	server := startServer(t, false)
	first := receiptFrom("Bob <bob@example.com>")
	second := receiptFrom("alice@example.com")
	server.deliver(t, first)
	server.deliver(t, second)

	mailbox := connect(t, server, false)

	headers, err := mailbox.FetchHeaders([]uint32{2, 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 2 || headers[0].UID != 1 || headers[1].UID != 2 {
		t.Fatalf("FetchHeaders() = %+v", headers)
	}
	if headers[0].Size != int64(len(first)) {
		t.Errorf("Size = %d, want %d", headers[0].Size, len(first))
	}

	fields := ParseHeaderFields(headers[0].Header)
	if !slices.Equal(fields.From, []string{"bob@example.com"}) || fields.Subject != "Receipt" {
		t.Errorf("ParseHeaderFields() = %+v", fields)
	}
	if strings.Contains(string(headers[0].Header), "Total") {
		t.Error("the header fetch returned the body")
	}

	unseen, _ := mailbox.UnseenUIDs()
	if len(unseen) != 2 {
		t.Errorf("fetching headers marked messages read: UnseenUIDs() = %v", unseen)
	}
}

func TestMailbox_FetchHeadersOfNothing(t *testing.T) {
	server := startServer(t, false)
	mailbox := connect(t, server, false)

	headers, err := mailbox.FetchHeaders(nil)
	if err != nil || len(headers) != 0 {
		t.Errorf("FetchHeaders(nil) = %v, %v", headers, err)
	}
	if err := mailbox.MarkSeen(); err != nil {
		t.Errorf("MarkSeen() with no uids = %v", err)
	}
}

func TestConnector_CheckConnection(t *testing.T) {
	server := startServer(t, false)

	if err := server.connector().CheckConnection(server.credentials(false)); err != nil {
		t.Errorf("CheckConnection() = %v, want nil", err)
	}

	badPassword := server.credentials(false)
	badPassword.Password = "wrong"
	err := server.connector().CheckConnection(badPassword)
	if err == nil {
		t.Fatal("CheckConnection() with a wrong password = nil, want an error")
	}
	if strings.Contains(err.Error(), "wrong") {
		t.Errorf("the error leaks the password: %v", err)
	}
}

func TestConnector_RejectsAnUntrustedCertificate(t *testing.T) {
	server := startServer(t, false)

	// The zero Connector trusts only the system roots, like Python's default
	// SSL context, so a self-signed server certificate must fail.
	if err := (Connector{}).CheckConnection(server.credentials(false)); err == nil {
		t.Error("CheckConnection() trusted a self-signed certificate")
	}
}

func TestConnector_StartTLSRejectsAnUntrustedCertificate(t *testing.T) {
	server := startServer(t, true)

	if err := (Connector{}).CheckConnection(server.credentials(true)); err == nil {
		t.Error("CheckConnection() over STARTTLS trusted a self-signed certificate")
	}
}

func TestConnector_UnresponsiveServerTimesOut(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Accept and then say nothing, like a hung server.
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
		}
	}()

	host, port, _ := net.SplitHostPort(listener.Addr().String())
	connector := Connector{HandshakeTimeout: 200 * time.Millisecond}

	start := time.Now()
	err = connector.CheckConnection(Credentials{Host: host, Port: port, Username: "u", Password: "p"})
	if err == nil {
		t.Fatal("CheckConnection() = nil against a silent server")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("CheckConnection() took %v, want it bounded by the handshake timeout", elapsed)
	}
}

func TestConnector_UsesTheInjectedDialer(t *testing.T) {
	server := startServer(t, false)
	dialed := ""

	connector := server.connector()
	connector.Dial = func(network string, address string) (net.Conn, error) {
		dialed = address
		return net.Dial(network, address)
	}

	if err := connector.CheckConnection(server.credentials(false)); err != nil {
		t.Fatal(err)
	}
	if dialed != net.JoinHostPort(server.host, server.port) {
		t.Errorf("dialed %q", dialed)
	}
}

func TestMailbox_AddressesMessagesByUIDNotSequenceNumber(t *testing.T) {
	server := startServer(t, false)
	server.deliver(t, receiptFrom("first@example.com"))
	second := receiptFrom("second@example.com")
	third := receiptFrom("third@example.com")
	server.deliver(t, second)
	server.deliver(t, third)

	mailbox := connect(t, server, false)

	// Expunge the first message so UIDs 2 and 3 become sequence numbers 1 and 2.
	client := mailbox.(*imapMailbox).client
	if err := client.Store(toUIDSet([]uint32{1}), &imap.StoreFlags{Op: imap.StoreFlagsAdd, Silent: true, Flags: []imap.Flag{imap.FlagDeleted}}, nil).Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Expunge().Close(); err != nil {
		t.Fatal(err)
	}

	unseen, err := mailbox.UnseenUIDs()
	if err != nil || !slices.Equal(unseen, []uint32{2, 3}) {
		t.Fatalf("UnseenUIDs() = %v, %v; want [2 3]", unseen, err)
	}

	raw, err := mailbox.FetchRaw(2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, second) {
		t.Error("FetchRaw(2) did not return the message with UID 2")
	}

	if err := mailbox.MarkSeen(3); err != nil {
		t.Fatal(err)
	}
	unseen, _ = mailbox.UnseenUIDs()
	if !slices.Equal(unseen, []uint32{2}) {
		t.Errorf("after MarkSeen(3), UnseenUIDs() = %v, want [2]", unseen)
	}
}

func TestMailbox_SurvivesAPauseLongerThanTheHandshakeTimeout(t *testing.T) {
	server := startServer(t, false)
	server.deliver(t, receiptFrom("bob@example.com"))

	// imapclient waits for the next response with no deadline at all. A deadline
	// pushed onto every read, as the first port did, closes the session once the
	// poller spends long enough between two commands — converting an attachment,
	// say. imapclient retries a timed-out read up to four times at a response
	// boundary, so the pause has to outlast four deadlines to show it.
	connector := server.connector()
	connector.HandshakeTimeout = 100 * time.Millisecond
	mailbox, err := connector.Connect(server.credentials(false))
	if err != nil {
		t.Fatal(err)
	}
	defer mailbox.Close()

	time.Sleep(time.Second)

	unseen, err := mailbox.UnseenUIDs()
	if err != nil {
		t.Fatalf("UnseenUIDs() after a pause = %v, want the session still open", err)
	}
	if !slices.Equal(unseen, []uint32{1}) {
		t.Errorf("UnseenUIDs() = %v, want [1]", unseen)
	}
}

func TestMailbox_FetchRawOfAnExpungedMessageIsGone(t *testing.T) {
	server := startServer(t, false)
	server.deliver(t, receiptFrom("first@example.com"))
	second := receiptFrom("second@example.com")
	server.deliver(t, second)

	mailbox := connect(t, server, false)
	if _, err := mailbox.UnseenUIDs(); err != nil {
		t.Fatal(err)
	}

	// Another client deletes the message after it was listed.
	client := mailbox.(*imapMailbox).client
	if err := client.Store(toUIDSet([]uint32{1}), &imap.StoreFlags{Op: imap.StoreFlagsAdd, Silent: true, Flags: []imap.Flag{imap.FlagDeleted}}, nil).Close(); err != nil {
		t.Fatal(err)
	}
	if err := client.Expunge().Close(); err != nil {
		t.Fatal(err)
	}

	_, err := mailbox.FetchRaw(1)
	var fetchErr *MessageFetchError
	if !errors.As(err, &fetchErr) || !fetchErr.Gone || fetchErr.Temporary {
		t.Fatalf("FetchRaw() error = %#v, want a MessageFetchError marked Gone", err)
	}
	if !strings.Contains(err.Error(), "message 1") {
		t.Errorf("Error() = %q, want the UID named", err.Error())
	}

	raw, err := mailbox.FetchRaw(2)
	if err != nil || !bytes.Equal(raw, second) {
		t.Errorf("FetchRaw(2) after a vanished message = %v; the session should still work", err)
	}
}

// refusingSession answers every FETCH of one UID with the given status. It
// embeds the IMAP4rev2 session because the test server advertises that.
type refusingSession struct {
	imapserver.SessionIMAP4rev2
	uid     imap.UID
	refusal *imap.Error
}

func (session refusingSession) Fetch(w *imapserver.FetchWriter, numSet imap.NumSet, options *imap.FetchOptions) error {
	if uidSet, ok := numSet.(imap.UIDSet); ok && uidSet.Contains(session.uid) {
		return session.refusal
	}
	return session.SessionIMAP4rev2.Fetch(w, numSet, options)
}

func TestMailbox_FetchRawRefusals(t *testing.T) {
	cases := []struct {
		name          string
		refusal       *imap.Error
		wantMessage   bool
		wantTemporary bool
	}{
		{
			name:        "NO is about the message",
			refusal:     &imap.Error{Type: imap.StatusResponseTypeNo, Text: "message is damaged"},
			wantMessage: true,
		},
		{
			name:          "NO [UNAVAILABLE] is worth retrying",
			refusal:       &imap.Error{Type: imap.StatusResponseTypeNo, Code: imap.ResponseCodeUnavailable, Text: "try again later"},
			wantMessage:   true,
			wantTemporary: true,
		},
		{
			name:    "BAD is about the session",
			refusal: &imap.Error{Type: imap.StatusResponseTypeBad, Text: "syntax error"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := startServerWith(t, false, func(session imapserver.Session) imapserver.Session {
				return refusingSession{SessionIMAP4rev2: session.(imapserver.SessionIMAP4rev2), uid: 1, refusal: tc.refusal}
			})
			server.deliver(t, receiptFrom("refused@example.com"))
			mailbox := connect(t, server, false)

			_, err := mailbox.FetchRaw(1)
			if err == nil {
				t.Fatal("FetchRaw() = nil error for a refused message")
			}

			var fetchErr *MessageFetchError
			isMessageError := errors.As(err, &fetchErr)
			if isMessageError != tc.wantMessage {
				t.Fatalf("FetchRaw() error = %#v; a MessageFetchError = %v, want %v", err, isMessageError, tc.wantMessage)
			}
			if isMessageError && (fetchErr.Gone || fetchErr.Temporary != tc.wantTemporary) {
				t.Errorf("MessageFetchError = %+v, want Temporary %v and not Gone", fetchErr, tc.wantTemporary)
			}
			// The server's own status stays reachable through the wrapper.
			var imapErr *imap.Error
			if !errors.As(err, &imapErr) || imapErr.Text != tc.refusal.Text {
				t.Errorf("FetchRaw() error = %v, want it to unwrap to the server's %q", err, tc.refusal.Text)
			}
		})
	}
}
