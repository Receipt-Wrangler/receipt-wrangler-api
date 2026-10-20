package wranglerasynq

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/emersion/go-msgauth/dkim"
	"gorm.io/gorm"
	"receipt-wrangler/api/internal/emailclient"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
)

const (
	pollTestEncryptionKey = "email-poll-test-encryption-key"
	pollTestPassword      = "mailbox secret"
	pollTestSelector      = "test"
)

// --- Fake mailbox -----------------------------------------------------------------

type fakeMessage struct {
	raw  []byte
	size int64
	seen bool
}

// fakeMailbox is an in-memory emailclient.Mailbox that records what was asked of it.
type fakeMailbox struct {
	mu          sync.Mutex
	messages    map[uint32]*fakeMessage
	nextUID     uint32
	rawFetches  []uint32
	unmarked    []uint32
	connectErr  error
	fetchRawErr error
	// fetchRawErrs fails FetchRaw for one message only.
	fetchRawErrs  map[uint32]error
	markSeenErr   error
	markUnseenErr error
}

func newFakeMailbox() *fakeMailbox {
	return &fakeMailbox{messages: make(map[uint32]*fakeMessage), nextUID: 1}
}

// deliver adds an unread message and returns its UID.
func (mailbox *fakeMailbox) deliver(raw []byte) uint32 {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	uid := mailbox.nextUID
	mailbox.nextUID++
	mailbox.messages[uid] = &fakeMessage{raw: raw, size: int64(len(raw))}
	return uid
}

func (mailbox *fakeMailbox) isSeen(uid uint32) bool {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	return mailbox.messages[uid].seen
}

func (mailbox *fakeMailbox) UnseenUIDs() ([]uint32, error) {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	uids := make([]uint32, 0)
	for uid, message := range mailbox.messages {
		if !message.seen {
			uids = append(uids, uid)
		}
	}
	slices.Sort(uids)
	return uids, nil
}

func (mailbox *fakeMailbox) FetchHeaders(uids []uint32) ([]emailclient.MessageHeader, error) {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	headers := make([]emailclient.MessageHeader, 0, len(uids))
	for _, uid := range uids {
		message := mailbox.messages[uid]
		end := bytes.Index(message.raw, []byte("\r\n\r\n"))
		if end < 0 {
			end = len(message.raw)
		}
		headers = append(headers, emailclient.MessageHeader{UID: uid, Size: message.size, Header: message.raw[:end]})
	}
	return headers, nil
}

func (mailbox *fakeMailbox) FetchRaw(uid uint32) ([]byte, error) {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	mailbox.rawFetches = append(mailbox.rawFetches, uid)
	if mailbox.fetchRawErr != nil {
		return nil, mailbox.fetchRawErr
	}
	if err := mailbox.fetchRawErrs[uid]; err != nil {
		return nil, err
	}
	return mailbox.messages[uid].raw, nil
}

func (mailbox *fakeMailbox) MarkSeen(uids ...uint32) error {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	if mailbox.markSeenErr != nil {
		return mailbox.markSeenErr
	}
	for _, uid := range uids {
		mailbox.messages[uid].seen = true
	}
	return nil
}

func (mailbox *fakeMailbox) MarkUnseen(uids ...uint32) error {
	mailbox.mu.Lock()
	defer mailbox.mu.Unlock()
	mailbox.unmarked = append(mailbox.unmarked, uids...)
	if mailbox.markUnseenErr != nil {
		return mailbox.markUnseenErr
	}
	for _, uid := range uids {
		mailbox.messages[uid].seen = false
	}
	return nil
}

func (mailbox *fakeMailbox) Close() error { return nil }

// --- Fixture ------------------------------------------------------------------------

type pollFixture struct {
	t          *testing.T
	db         *gorm.DB
	mailboxes  map[string]*fakeMailbox
	connects   []emailclient.Credentials
	enqueued   []EmailProcessTaskPayload
	enqueueErr error
	// failEnqueueCall fails only that enqueue call, counting from 1 across polls.
	failEnqueueCall int
	enqueueCalls    int
	lookupTXT       emailclient.TXTLookup
}

func newPollFixture(t *testing.T) *pollFixture {
	t.Setenv("ENCRYPTION_KEY", pollTestEncryptionKey)
	t.Setenv("BASE_PATH", t.TempDir())
	t.Cleanup(repositories.TruncateTestDb)

	fixture := &pollFixture{
		t:         t,
		db:        repositories.GetDB(),
		mailboxes: make(map[string]*fakeMailbox),
		lookupTXT: func(name string) ([]string, error) {
			return nil, &net.DNSError{Err: "no such host", Name: name, IsNotFound: true}
		},
	}

	original := newEmailPoller
	newEmailPoller = fixture.poller
	t.Cleanup(func() { newEmailPoller = original })

	return fixture
}

func (fixture *pollFixture) poller() emailPoller {
	return emailPoller{
		connect: func(credentials emailclient.Credentials) (emailclient.Mailbox, error) {
			fixture.connects = append(fixture.connects, credentials)
			mailbox, ok := fixture.mailboxes[credentials.Username]
			if !ok {
				return nil, fmt.Errorf("no mailbox for %s", credentials.Username)
			}
			if mailbox.connectErr != nil {
				return nil, mailbox.connectErr
			}
			return mailbox, nil
		},
		lookupTXT: func(name string) ([]string, error) { return fixture.lookupTXT(name) },
		enqueue: func(payload EmailProcessTaskPayload) error {
			fixture.enqueueCalls++
			if fixture.enqueueErr != nil {
				return fixture.enqueueErr
			}
			if fixture.enqueueCalls == fixture.failEnqueueCall {
				return errors.New("redis unavailable")
			}
			fixture.enqueued = append(fixture.enqueued, payload)
			return nil
		},
	}
}

func (fixture *pollFixture) addSystemEmail(username string) (models.SystemEmail, *fakeMailbox) {
	fixture.t.Helper()
	password, err := utils.EncryptAndEncodeToBase64(pollTestEncryptionKey, pollTestPassword)
	if err != nil {
		fixture.t.Fatal(err)
	}

	systemEmail := models.SystemEmail{Host: "imap.example.org", Port: "993", Username: username, Password: password}
	if err := fixture.db.Create(&systemEmail).Error; err != nil {
		fixture.t.Fatal(err)
	}

	mailbox := newFakeMailbox()
	fixture.mailboxes[username] = mailbox
	return systemEmail, mailbox
}

type groupOptions struct {
	whitelist      []string
	regexes        []string
	bodyProcessing bool
	requireSigned  bool
	disabled       bool
}

func (fixture *pollFixture) addGroup(systemEmail models.SystemEmail, options groupOptions) models.GroupSettings {
	fixture.t.Helper()
	group := models.Group{Name: "group"}
	if err := fixture.db.Create(&group).Error; err != nil {
		fixture.t.Fatal(err)
	}

	settings := models.GroupSettings{
		GroupId:                    group.ID,
		EmailIntegrationEnabled:    !options.disabled,
		EmailBodyProcessingEnabled: options.bodyProcessing,
		SystemEmailId:              &systemEmail.ID,
		EmailDefaultReceiptStatus:  models.OPEN,
	}
	for _, email := range options.whitelist {
		settings.EmailWhiteList = append(settings.EmailWhiteList, models.GroupSettingsWhiteListEmail{Email: email})
	}
	for _, regex := range options.regexes {
		settings.SubjectLineRegexes = append(settings.SubjectLineRegexes, models.SubjectLineRegex{Regex: regex})
	}
	if err := fixture.db.Create(&settings).Error; err != nil {
		fixture.t.Fatal(err)
	}
	// The column defaults to true and GORM skips a false bool on create.
	if err := fixture.db.Model(&settings).Update("email_require_signed_sender", options.requireSigned).Error; err != nil {
		fixture.t.Fatal(err)
	}

	loaded, err := repositories.NewGroupSettingsRepository(nil).GetGroupSettingsById(utils.UintToString(settings.ID))
	if err != nil {
		fixture.t.Fatal(err)
	}
	return loaded
}

func (fixture *pollFixture) enabledGroupSettings() []models.GroupSettings {
	fixture.t.Helper()
	settings, err := repositories.NewGroupSettingsRepository(nil).GetAllGroupSettings("email_integration_enabled = ?", true)
	if err != nil {
		fixture.t.Fatal(err)
	}
	return settings
}

func (fixture *pollFixture) payloadsFor(groupSettingsId uint) []EmailProcessTaskPayload {
	result := make([]EmailProcessTaskPayload, 0)
	for _, payload := range fixture.enqueued {
		if payload.GroupSettingsId == groupSettingsId {
			result = append(result, payload)
		}
	}
	return result
}

func (fixture *pollFixture) systemTasks(taskType models.SystemTaskType) []models.SystemTask {
	fixture.t.Helper()
	var tasks []models.SystemTask
	if err := fixture.db.Where("type = ?", taskType).Order("id").Find(&tasks).Error; err != nil {
		fixture.t.Fatal(err)
	}
	return tasks
}

// --- Messages -------------------------------------------------------------------------

var (
	pollJpeg = append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("JFIF photo of a receipt")...)
	// A PDF header with no PDF behind it: sniffed as a PDF, fails conversion.
	corruptPdf = []byte("%PDF-1.4\nthis is not really a pdf\n")
)

type fixtureAttachment struct {
	filename    string
	contentType string
	data        []byte
}

func jpegNamed(filename string, data []byte) fixtureAttachment {
	return fixtureAttachment{filename: filename, contentType: "image/jpeg", data: data}
}

// rawEmail builds a multipart/mixed message with an optional plain or HTML body.
func rawEmail(from string, subject string, body string, attachments ...fixtureAttachment) []byte {
	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "From: %s\r\nTo: receipts@example.org\r\nSubject: %s\r\nDate: Mon, 01 Jan 2024 22:00:00 -0500\r\nMIME-Version: 1.0\r\n", from, subject)
	buffer.WriteString("Content-Type: multipart/mixed; boundary=\"b\"\r\n\r\n")

	if len(body) > 0 {
		contentType := "text/plain"
		if strings.HasPrefix(body, "<") {
			contentType = "text/html"
		}
		fmt.Fprintf(&buffer, "--b\r\nContent-Type: %s; charset=utf-8\r\n\r\n%s\r\n", contentType, body)
	}

	for _, attachment := range attachments {
		fmt.Fprintf(&buffer, "--b\r\nContent-Type: %s\r\nContent-Disposition: attachment; filename=%q\r\nContent-Transfer-Encoding: base64\r\n\r\n%s\r\n",
			attachment.contentType, attachment.filename, base64.StdEncoding.EncodeToString(attachment.data))
	}

	buffer.WriteString("--b--\r\n")
	return buffer.Bytes()
}

// --- DKIM -------------------------------------------------------------------------------

var (
	pollKeyOnce sync.Once
	pollKey     *rsa.PrivateKey
)

func pollSigningKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	pollKeyOnce.Do(func() {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		pollKey = key
	})
	return pollKey
}

func signEmail(t *testing.T, raw []byte, domain string) []byte {
	t.Helper()
	var signed bytes.Buffer
	err := dkim.Sign(&signed, bytes.NewReader(raw), &dkim.SignOptions{Domain: domain, Selector: pollTestSelector, Signer: pollSigningKey(t)})
	if err != nil {
		t.Fatal(err)
	}
	return signed.Bytes()
}

// publishKey makes the fixture's DNS serve the signing key for domain.
func (fixture *pollFixture) publishKey(domain string) {
	der, err := x509.MarshalPKIXPublicKey(&pollSigningKey(fixture.t).PublicKey)
	if err != nil {
		fixture.t.Fatal(err)
	}
	record := "v=DKIM1; k=rsa; p=" + base64.StdEncoding.EncodeToString(der)
	fixture.lookupTXT = func(name string) ([]string, error) {
		if name == pollTestSelector+"._domainkey."+domain {
			return []string{record}, nil
		}
		return nil, &net.DNSError{Err: "no such host", Name: name, IsNotFound: true}
	}
}

var errDNSDown = errors.New("dns down")

func (fixture *pollFixture) dnsTemporarilyDown() {
	fixture.lookupTXT = func(name string) ([]string, error) {
		return nil, &net.DNSError{Err: errDNSDown.Error(), Name: name, IsTemporary: true}
	}
}
