package emailclient

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/emersion/go-msgauth/dkim"
	"golang.org/x/net/publicsuffix"
)

// VerdictKind is the outcome of checking a sender's DKIM signature.
type VerdictKind int

const (
	// Authenticated means a valid signature from the sender's own domain.
	Authenticated VerdictKind = iota
	// Unauthenticated means no such signature: missing, invalid, or from
	// another domain.
	Unauthenticated
	// RetryLater means a signature from the sender's domain could not be
	// checked for a temporary reason, such as a DNS failure.
	RetryLater
)

// Verdict is a VerdictKind with a human-readable reason for anything but a pass.
type Verdict struct {
	Kind   VerdictKind
	Reason string
}

// TXTLookup resolves DNS TXT records. net.LookupTXT in production.
type TXTLookup func(name string) ([]string, error)

// VerifySender checks that the message carries a valid DKIM signature from the
// organizational domain of fromAddress.
//
// The library does the cryptography and already refuses signatures that do not
// cover From, use the insecure l= tag, use SHA-1, or use an RSA key under 1024
// bits. What it does not decide is whose signature counts: a message validly
// signed by a bulk-mail service's domain is a clean pass to the library, so the
// alignment with the From domain is checked here.
func VerifySender(raw []byte, fromAddress string, lookup TXTLookup) Verdict {
	fromDomain := domainOf(fromAddress)
	if len(fromDomain) == 0 {
		return Verdict{Kind: Unauthenticated, Reason: "the sender address has no domain"}
	}

	if lookup == nil {
		lookup = net.LookupTXT
	}

	verifications, err := dkim.VerifyWithOptions(bytes.NewReader(raw), &dkim.VerifyOptions{LookupTXT: lookup})
	if err != nil {
		return Verdict{Kind: Unauthenticated, Reason: fmt.Sprintf("the message could not be read for DKIM verification: %s", err.Error())}
	}

	if len(verifications) == 0 {
		return Verdict{Kind: Unauthenticated, Reason: fmt.Sprintf("the message has no DKIM signature from %s", fromDomain)}
	}

	retry := false
	var reasons []string

	for _, verification := range verifications {
		aligned := SameOrganizationalDomain(verification.Domain, fromDomain)

		switch {
		case verification.Err == nil && aligned:
			return Verdict{Kind: Authenticated}
		case verification.Err == nil:
			reasons = append(reasons, fmt.Sprintf("signed by %s, not %s", verification.Domain, fromDomain))
		case dkim.IsTempFail(verification.Err) && aligned:
			// Only worth retrying when the signature, once checkable, would count.
			retry = true
		default:
			reasons = append(reasons, fmt.Sprintf("signature from %s is invalid: %s", verification.Domain, verification.Err.Error()))
		}
	}

	if retry {
		return Verdict{Kind: RetryLater, Reason: fmt.Sprintf("the DKIM signature from %s could not be checked right now", fromDomain)}
	}

	return Verdict{Kind: Unauthenticated, Reason: "no valid DKIM signature from " + fromDomain + " (" + strings.Join(reasons, "; ") + ")"}
}

// SameOrganizationalDomain reports whether two domains belong to the same
// registered domain, e.g. mail.example.com and example.com, the relaxed alignment
// DMARC uses. A bare public suffix such as co.uk never matches.
func SameOrganizationalDomain(a string, b string) bool {
	a = normalizeDomain(a)
	b = normalizeDomain(b)
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	organizationA, err := publicsuffix.EffectiveTLDPlusOne(a)
	if err != nil {
		return false
	}
	organizationB, err := publicsuffix.EffectiveTLDPlusOne(b)
	if err != nil {
		return false
	}

	return organizationA == organizationB
}

func domainOf(address string) string {
	index := strings.LastIndex(address, "@")
	if index < 0 {
		return ""
	}
	return normalizeDomain(address[index+1:])
}

func normalizeDomain(domain string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(domain)), ".")
}
