package emailclient

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/emersion/go-msgauth/dkim"
)

const testSelector = "test"

var (
	testKeyOnce sync.Once
	testKey     *rsa.PrivateKey
	otherKey    *rsa.PrivateKey
)

func signingKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PrivateKey) {
	t.Helper()
	testKeyOnce.Do(func() {
		var err error
		if testKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
			panic(err)
		}
		if otherKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
			panic(err)
		}
	})
	return testKey, otherKey
}

// dnsRecord is the TXT record publishing key's public half.
func dnsRecord(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return "v=DKIM1; k=rsa; p=" + base64.StdEncoding.EncodeToString(der)
}

// publishedKeys returns a lookup serving testKey for every listed domain.
func publishedKeys(t *testing.T, domains ...string) TXTLookup {
	key, _ := signingKeys(t)
	record := dnsRecord(t, key)
	return func(name string) ([]string, error) {
		for _, domain := range domains {
			if name == testSelector+"._domainkey."+domain {
				return []string{record}, nil
			}
		}
		return nil, &net.DNSError{Err: "no such host", Name: name, IsNotFound: true}
	}
}

func sign(t *testing.T, raw []byte, domain string, key *rsa.PrivateKey) []byte {
	t.Helper()
	var signed bytes.Buffer
	err := dkim.Sign(&signed, bytes.NewReader(raw), &dkim.SignOptions{
		Domain:   domain,
		Selector: testSelector,
		Signer:   key,
	})
	if err != nil {
		t.Fatalf("dkim.Sign() error = %v", err)
	}
	return signed.Bytes()
}

func receiptFrom(from string) []byte {
	return buildMessage([]string{
		"From: " + from,
		"To: receipts@example.org",
		"Subject: Receipt",
		"Date: Wed, 20 Oct 2021 10:30:00 +0000",
	}, textPart("text/plain; charset=utf-8", "Total $12.00\r\n"))
}

func TestVerifySender(t *testing.T) {
	key, other := signingKeys(t)

	cases := []struct {
		name   string
		raw    func() []byte
		from   string
		lookup TXTLookup
		want   VerdictKind
	}{
		{
			name:   "aligned signature",
			raw:    func() []byte { return sign(t, receiptFrom("bob@example.com"), "example.com", key) },
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Authenticated,
		},
		{
			name:   "subdomain sender signed by parent domain",
			raw:    func() []byte { return sign(t, receiptFrom("noreply@mail.example.com"), "example.com", key) },
			from:   "noreply@mail.example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Authenticated,
		},
		{
			name:   "mixed case domains",
			raw:    func() []byte { return sign(t, receiptFrom("Bob@EXAMPLE.com"), "Example.COM", key) },
			from:   "Bob@EXAMPLE.com",
			lookup: publishedKeys(t, "Example.COM"),
			want:   Authenticated,
		},
		{
			name:   "valid signature from another domain only",
			raw:    func() []byte { return sign(t, receiptFrom("bob@example.com"), "sendgrid.net", key) },
			from:   "bob@example.com",
			lookup: publishedKeys(t, "sendgrid.net"),
			want:   Unauthenticated,
		},
		{
			name: "third-party signature plus aligned signature",
			raw: func() []byte {
				return sign(t, sign(t, receiptFrom("bob@example.com"), "example.com", key), "sendgrid.net", key)
			},
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com", "sendgrid.net"),
			want:   Authenticated,
		},
		{
			name:   "unsigned",
			raw:    func() []byte { return receiptFrom("bob@example.com") },
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Unauthenticated,
		},
		{
			name: "body altered after signing",
			raw: func() []byte {
				signed := sign(t, receiptFrom("bob@example.com"), "example.com", key)
				return bytes.Replace(signed, []byte("Total $12.00"), []byte("Total $99.00"), 1)
			},
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Unauthenticated,
		},
		{
			name: "subject altered after signing",
			raw: func() []byte {
				signed := sign(t, receiptFrom("bob@example.com"), "example.com", key)
				return bytes.Replace(signed, []byte("Subject: Receipt"), []byte("Subject: Refund"), 1)
			},
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Unauthenticated,
		},
		{
			name:   "signed with a key the domain did not publish",
			raw:    func() []byte { return sign(t, receiptFrom("bob@example.com"), "example.com", other) },
			from:   "bob@example.com",
			lookup: publishedKeys(t, "example.com"),
			want:   Unauthenticated,
		},
		{
			name: "public suffix is not an organization",
			raw:  func() []byte { return sign(t, receiptFrom("bob@shop.co.uk"), "co.uk", key) },
			from: "bob@shop.co.uk",
			lookup: func(name string) ([]string, error) {
				return []string{dnsRecord(t, key)}, nil
			},
			want: Unauthenticated,
		},
		{
			name: "dns temporarily unavailable for the aligned signature",
			raw:  func() []byte { return sign(t, receiptFrom("bob@example.com"), "example.com", key) },
			from: "bob@example.com",
			lookup: func(name string) ([]string, error) {
				return nil, &net.DNSError{Err: "server misbehaving", Name: name, IsTemporary: true}
			},
			want: RetryLater,
		},
		{
			name: "dns temporarily unavailable for a third-party signature only",
			raw:  func() []byte { return sign(t, receiptFrom("bob@example.com"), "sendgrid.net", key) },
			from: "bob@example.com",
			lookup: func(name string) ([]string, error) {
				return nil, &net.DNSError{Err: "server misbehaving", Name: name, IsTemporary: true}
			},
			// Retrying cannot help: that signature would not count anyway.
			want: Unauthenticated,
		},
		{
			name:   "sender without a domain",
			raw:    func() []byte { return receiptFrom("bob@example.com") },
			from:   "bob",
			lookup: publishedKeys(t),
			want:   Unauthenticated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict := VerifySender(tc.raw(), tc.from, tc.lookup)
			if verdict.Kind != tc.want {
				t.Errorf("VerifySender() = %+v, want kind %v", verdict, tc.want)
			}
			if verdict.Kind != Authenticated && len(verdict.Reason) == 0 {
				t.Error("a non-pass verdict carries no reason")
			}
		})
	}
}

func TestVerifySender_SecondFromHeaderIsNotSigned(t *testing.T) {
	// The library verifies the bottom-most From only. Adding a forged From above
	// it leaves the signature valid, which is why ParseMessage refuses messages
	// with more than one From before DKIM is ever consulted.
	key, _ := signingKeys(t)
	signed := sign(t, receiptFrom("bob@example.com"), "example.com", key)
	forged := append([]byte("From: ceo@victim.example\r\n"), signed...)

	if verdict := VerifySender(forged, "bob@example.com", publishedKeys(t, "example.com")); verdict.Kind != Authenticated {
		t.Fatalf("precondition: the signature still verifies, got %+v", verdict)
	}
	if _, err := ParseMessage(forged); !errors.Is(err, ErrAmbiguousFrom) {
		t.Errorf("ParseMessage() error = %v, want ErrAmbiguousFrom", err)
	}
}

func TestSameOrganizationalDomain(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"example.com", "example.com", true},
		{"mail.example.com", "example.com", true},
		{"a.example.com", "b.example.com", true},
		{"EXAMPLE.com.", "example.COM", true},
		{"example.com", "example.org", false},
		{"example.co.uk", "other.co.uk", false},
		{"co.uk", "shop.co.uk", false},
		{"", "example.com", false},
	}

	for _, tc := range cases {
		t.Run(strings.Join([]string{tc.a, tc.b}, "|"), func(t *testing.T) {
			if got := SameOrganizationalDomain(tc.a, tc.b); got != tc.want {
				t.Errorf("SameOrganizationalDomain(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
