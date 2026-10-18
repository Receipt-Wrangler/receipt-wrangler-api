package emailclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-message"
	// Registers the decoders for non-UTF-8 charsets (windows-1252, ISO-8859-x,
	// Shift_JIS, ...). Without it every such part reads as an unknown charset.
	_ "github.com/emersion/go-message/charset"
	gomail "github.com/emersion/go-message/mail"
	"golang.org/x/net/html"

	"receipt-wrangler/api/internal/utils"
)

var (
	// ErrMissingFrom is returned for a message without a parseable From address.
	ErrMissingFrom = errors.New("message has no sender address")

	// ErrAmbiguousFrom is returned for a message carrying more than one From
	// header or more than one address in it. A DKIM signature covers only one
	// From header, so a second one is how a forged sender rides alongside a
	// genuine signature: the verifier checks one, a naive reader uses the other.
	ErrAmbiguousFrom = errors.New("message has more than one sender address")
)

// InlineImageMaxBytes is the size below which an inline image embedded in the
// HTML body (referenced as cid:) is treated as decoration — a signature logo or a
// tracking pixel — rather than a receipt. Photos attached from Apple Mail and iOS
// Mail are inline too, but are far larger.
const InlineImageMaxBytes = 20 * 1024

// attachmentMimeTypes is the set of sniffed types an email attachment may have to
// become a receipt. It is the list the Python client accepted, now matched
// against the detected type rather than the sender-declared Content-Type.
var attachmentMimeTypes = regexp.MustCompile(`^(image/(jpeg|png|heic|bmp|webp|tiff)|application/pdf)$`)

var (
	horizontalWhitespace = regexp.MustCompile(`[ \t]+`)
	excessiveNewlines    = regexp.MustCompile(`\n{3,}`)
)

// Address is one parsed mailbox address.
type Address struct {
	Name  string
	Email string
}

// Attachment is one attachment that passed the type allowlist.
type Attachment struct {
	// Filename is the sender's name for the file, reduced to a safe basename.
	Filename string
	// FileType is the detected MIME type of Data.
	FileType string
	Data     []byte
}

// Message is the part of an email the receipt pipeline uses.
type Message struct {
	From        Address
	To          string
	Subject     string
	Date        time.Time
	Body        string
	BodyHtml    string
	Attachments []Attachment
}

// HasContent reports whether there is anything to make a receipt from.
func (message Message) HasContent() bool {
	return len(message.Attachments) > 0 || len(message.Body) > 0 || len(message.BodyHtml) > 0
}

// HeaderFields are the fields routing needs, read before the body is fetched.
type HeaderFields struct {
	// From holds every address found in every From header. Normally one; a
	// message with more is rejected once its full body is parsed.
	From    []string
	Subject string
}

// ParseHeaderFields reads From and Subject from a raw header block. It never
// fails: an unparseable From yields no addresses.
func ParseHeaderFields(rawHeader []byte) HeaderFields {
	header, err := readHeader(rawHeader)
	if err != nil {
		return HeaderFields{}
	}

	fields := HeaderFields{Subject: subjectOf(header)}
	for _, value := range header.Values("From") {
		addresses, err := parseAddressList(value)
		if err != nil {
			continue
		}
		for _, address := range addresses {
			fields.From = append(fields.From, address.Email)
		}
	}

	return fields
}

// ParseMessage parses a raw RFC 5322 message.
func ParseMessage(raw []byte) (Message, error) {
	reader, err := gomail.CreateReader(bytes.NewReader(raw))
	if err != nil && !message.IsUnknownCharset(err) {
		return Message{}, fmt.Errorf("could not parse message: %w", err)
	}
	defer reader.Close()

	from, err := singleFrom(reader.Header)
	if err != nil {
		return Message{}, err
	}

	result := Message{
		From:    from,
		To:      firstAddress(reader.Header, "To"),
		Subject: subjectOf(reader.Header),
		Date:    dateOf(reader.Header),
	}

	var plainParts, htmlParts []string
	var candidates []attachmentCandidate

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil && !message.IsUnknownCharset(err) {
			if message.IsUnknownEncoding(err) {
				// The part's transfer encoding is unknown, so its bytes are
				// meaningless. Skip it rather than lose the whole email.
				continue
			}
			return Message{}, fmt.Errorf("could not parse message part: %w", err)
		}
		unknownCharset := err != nil

		header := partHeader(part)
		data, readErr := io.ReadAll(part.Body)
		if readErr != nil {
			return Message{}, fmt.Errorf("could not read message part: %w", readErr)
		}

		if kind := classifyBodyPart(header); kind != notBody {
			// An undecodable or empty body part is neither body text nor an
			// attachment.
			if !unknownCharset && len(data) > 0 {
				text := strings.ToValidUTF8(string(data), "�")
				if kind == plainBody {
					plainParts = append(plainParts, text)
				} else {
					htmlParts = append(htmlParts, text)
				}
			}
			continue
		}

		if candidate, ok := newAttachmentCandidate(header, data); ok {
			candidates = append(candidates, candidate)
		}
	}

	result.BodyHtml = strings.Join(htmlParts, "\n")
	result.Body = formatBodyText(plainParts, htmlParts)
	result.Attachments = keepReceiptAttachments(candidates, result.BodyHtml)

	return result, nil
}

// attachmentCandidate is a part that passed the type allowlist, waiting on the
// inline-image rule, which needs the HTML body.
type attachmentCandidate struct {
	attachment Attachment
	inline     bool
	contentId  string
}

func newAttachmentCandidate(header message.Header, data []byte) (attachmentCandidate, bool) {
	// Like the Python client, a part with no Content-Disposition is never an
	// attachment. Those are typically images an HTML body embeds by cid.
	rawDisposition := header.Get("Content-Disposition")
	if len(rawDisposition) == 0 {
		return attachmentCandidate{}, false
	}

	filename := utils.SanitizeFileName(partFilename(header))
	if len(filename) == 0 {
		return attachmentCandidate{}, false
	}

	fileType := utils.GetMimeType(data).String()
	if !attachmentMimeTypes.MatchString(fileType) {
		return attachmentCandidate{}, false
	}

	disposition, _, _ := header.ContentDisposition()

	return attachmentCandidate{
		attachment: Attachment{Filename: filename, FileType: fileType, Data: data},
		inline:     strings.EqualFold(disposition, "inline"),
		contentId:  strings.Trim(strings.TrimSpace(header.Get("Content-Id")), "<>"),
	}, true
}

// keepReceiptAttachments drops inline images that are HTML decoration: embedded
// by the body (cid:) and small. Everything else is kept, in message order.
func keepReceiptAttachments(candidates []attachmentCandidate, bodyHtml string) []Attachment {
	lowerHtml := strings.ToLower(bodyHtml)
	attachments := make([]Attachment, 0, len(candidates))

	for _, candidate := range candidates {
		embedded := candidate.inline &&
			len(candidate.contentId) > 0 &&
			strings.Contains(lowerHtml, "cid:"+strings.ToLower(candidate.contentId))

		if embedded && len(candidate.attachment.Data) < InlineImageMaxBytes {
			continue
		}

		attachments = append(attachments, candidate.attachment)
	}

	return attachments
}

type bodyKind int

const (
	notBody bodyKind = iota
	plainBody
	htmlBody
)

// classifyBodyPart mirrors the Python client: a text/plain or text/html part is
// body text unless its Content-Disposition mentions "attachment".
func classifyBodyPart(header message.Header) bodyKind {
	if strings.Contains(strings.ToLower(header.Get("Content-Disposition")), "attachment") {
		return notBody
	}

	contentType, _, _ := header.ContentType()
	switch strings.ToLower(contentType) {
	case "text/plain":
		return plainBody
	case "text/html":
		return htmlBody
	}
	return notBody
}

// formatBodyText prefers the plain-text parts and falls back to the HTML parts
// stripped of markup, then collapses whitespace the way the Python client did.
func formatBodyText(plainParts []string, htmlParts []string) string {
	var text string
	switch {
	case len(plainParts) > 0:
		text = strings.Join(plainParts, "\n")
	case len(htmlParts) > 0:
		stripped := make([]string, len(htmlParts))
		for i, part := range htmlParts {
			stripped[i] = StripHtmlTags(part)
		}
		text = strings.Join(stripped, "\n")
	default:
		return ""
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = horizontalWhitespace.ReplaceAllString(text, " ")
	text = excessiveNewlines.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

// StripHtmlTags returns the visible text of an HTML fragment: text nodes joined
// with nothing between them, entities decoded. The contents of <style> and
// <script> are not visible text and are dropped, so CSS and scripts never reach
// the AI prompt.
func StripHtmlTags(fragment string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(fragment))
	var builder strings.Builder
	skipDepth := 0

	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return builder.String()
		case html.StartTagToken:
			if isHiddenTextElement(tokenizer) {
				skipDepth++
			}
		case html.EndTagToken:
			if isHiddenTextElement(tokenizer) && skipDepth > 0 {
				skipDepth--
			}
		case html.TextToken:
			if skipDepth == 0 {
				builder.WriteString(tokenizer.Token().Data)
			}
		}
	}
}

func isHiddenTextElement(tokenizer *html.Tokenizer) bool {
	name, _ := tokenizer.TagName()
	tag := string(name)
	return tag == "style" || tag == "script"
}

func singleFrom(header gomail.Header) (Address, error) {
	values := header.Values("From")
	if len(values) == 0 {
		return Address{}, ErrMissingFrom
	}
	if len(values) > 1 {
		return Address{}, ErrAmbiguousFrom
	}

	addresses, err := parseAddressList(values[0])
	if err != nil || len(addresses) == 0 {
		return Address{}, ErrMissingFrom
	}
	if len(addresses) > 1 {
		return Address{}, ErrAmbiguousFrom
	}

	return addresses[0], nil
}

// parseAddressList parses one address-list header value, decoding encoded-word
// display names.
func parseAddressList(value string) ([]Address, error) {
	var header gomail.Header
	header.Set("From", value)

	parsed, err := header.AddressList("From")
	if err != nil {
		return nil, err
	}

	addresses := make([]Address, 0, len(parsed))
	for _, address := range parsed {
		addresses = append(addresses, Address{
			Name:  strings.TrimSpace(address.Name),
			Email: strings.TrimSpace(address.Address),
		})
	}

	return addresses, nil
}

func firstAddress(header gomail.Header, key string) string {
	addresses, err := header.AddressList(key)
	if err != nil || len(addresses) == 0 {
		return ""
	}
	return addresses[0].Address
}

func subjectOf(header gomail.Header) string {
	// On a decoding error Subject returns the raw value, which is still the best
	// available text.
	subject, _ := header.Subject()
	return subject
}

// dateOf parses the Date header and normalizes it to UTC. A missing or
// unparseable date is the zero time.
func dateOf(header gomail.Header) time.Time {
	value := strings.TrimSpace(header.Get("Date"))
	if len(value) == 0 {
		return time.Time{}
	}

	date, err := mail.ParseDate(value)
	if err != nil {
		// Some senders append a zone comment, e.g. "... +0000 (UTC)".
		if index := strings.Index(value, "("); index > 0 {
			date, err = mail.ParseDate(strings.TrimSpace(value[:index]))
		}
	}
	if err != nil {
		return time.Time{}
	}

	return date.UTC()
}

// partFilename reads the filename from Content-Disposition, falling back to the
// Content-Type name parameter, decoding RFC 2231 and RFC 2047 forms.
func partFilename(header message.Header) string {
	_, params, _ := header.ContentDisposition()
	if filename, ok := params["filename"]; ok && len(filename) > 0 {
		return filename
	}

	_, params, _ = header.ContentType()
	return params["name"]
}

func partHeader(part *gomail.Part) message.Header {
	switch header := part.Header.(type) {
	case *gomail.InlineHeader:
		return header.Header
	case *gomail.AttachmentHeader:
		return header.Header
	}
	return message.Header{}
}

func readHeader(rawHeader []byte) (gomail.Header, error) {
	entity, err := message.Read(bytes.NewReader(append(bytes.TrimRight(rawHeader, "\r\n"), []byte("\r\n\r\n")...)))
	if entity == nil {
		return gomail.Header{}, err
	}
	return gomail.Header{Header: entity.Header}, nil
}
