package emailclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"regexp"
	"slices"
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

// maxNestedMessageDepth bounds how far a forwarded message inside a forwarded
// message is followed. Python's walk() had no limit; this one keeps a crafted
// message from nesting arbitrarily deep.
const maxNestedMessageDepth = 5

// attachmentMimeTypes is the set of types an email attachment may have to become
// a receipt: the list the Python client accepted, plus HEIF. It is matched against
// the detected type, and against the declared Content-Type only to decide whether
// a part that failed was meant to be a receipt.
var attachmentMimeTypes = regexp.MustCompile(`^(image/(jpeg|png|heic|heif|bmp|webp|tiff)|application/pdf)$`)

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

// RejectedAttachment is an attachment that was meant to be a receipt but cannot
// become one: it would not decode, or its contents are not the image or PDF it
// was declared as.
type RejectedAttachment struct {
	Filename string
	Reason   string
}

// Message is the part of an email the receipt pipeline uses.
type Message struct {
	From                Address
	To                  string
	Subject             string
	Date                time.Time
	Body                string
	BodyHtml            string
	Attachments         []Attachment
	RejectedAttachments []RejectedAttachment
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

	var parts messageParts
	if err := parts.collect(reader, 0); err != nil {
		return Message{}, err
	}

	result.BodyHtml = strings.Join(parts.html, "\n")
	result.Body = formatBodyText(parts.plain, parts.html)
	result.Attachments, result.RejectedAttachments = keepReceiptAttachments(parts.candidates, result.BodyHtml)

	return result, nil
}

// messageParts collects the body text and attachment candidates of a message
// and of any message forwarded inside it.
type messageParts struct {
	plain      []string
	html       []string
	candidates []attachmentCandidate
}

func (parts *messageParts) collect(reader *gomail.Reader, depth int) error {
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return nil
		}
		if err != nil && !message.IsUnknownCharset(err) {
			if message.IsUnknownEncoding(err) {
				// The part's transfer encoding is unknown, so its bytes are
				// meaningless. Skip it rather than lose the whole email.
				continue
			}
			return fmt.Errorf("could not parse message part: %w", err)
		}
		unknownCharset := err != nil

		header := partHeader(part)
		data, readErr := io.ReadAll(part.Body)
		if readErr != nil {
			// One part that will not decode, such as broken base64, costs that
			// part only. The next NextPart skips whatever of it is left.
			if candidate, ok := newAttachmentCandidate(header, data, readErr); ok {
				parts.candidates = append(parts.candidates, candidate)
			}
			continue
		}

		if isNestedMessage(header) {
			parts.collectNested(data, depth+1)
			continue
		}

		if kind := classifyBodyPart(header); kind != notBody {
			// An undecodable or empty body part is neither body text nor an
			// attachment.
			if !unknownCharset && len(data) > 0 {
				text := strings.ToValidUTF8(string(data), "�")
				if kind == plainBody {
					parts.plain = append(parts.plain, text)
				} else {
					parts.html = append(parts.html, text)
				}
			}
			continue
		}

		if candidate, ok := newAttachmentCandidate(header, data, nil); ok {
			parts.candidates = append(parts.candidates, candidate)
		}
	}
}

// collectNested reads a message forwarded as an attachment, as Python's walk()
// did: its attachments become candidates and its text joins the body. Its
// headers are ignored. A forwarded message that cannot be read, or is nested too
// deep, is skipped without failing the email carrying it; whatever of it was read
// before a problem is kept.
func (parts *messageParts) collectNested(raw []byte, depth int) {
	if depth > maxNestedMessageDepth {
		return
	}

	reader, err := gomail.CreateReader(bytes.NewReader(raw))
	if err != nil && !message.IsUnknownCharset(err) {
		return
	}
	defer reader.Close()

	_ = parts.collect(reader, depth)
}

func isNestedMessage(header message.Header) bool {
	contentType, _ := contentTypeOf(header)
	return contentType == "message/rfc822" || contentType == "message/global"
}

// attachmentCandidate is a part meant as an attachment, waiting on the
// inline-image rule, which needs the HTML body. A rejected one carries why it
// cannot become a receipt instead of usable data.
type attachmentCandidate struct {
	attachment Attachment
	rejection  string
	inline     bool
	contentId  string
}

// newAttachmentCandidate decides what a part with a filename becomes. The sniffed
// type decides acceptance. The declared type only decides whether a part that
// fails was meant as a receipt, and so is reported rather than dropped: Python
// accepted a part on its declared type, so such a failure used to reach a group.
// readErr is the error reading the part, when it would not decode.
func newAttachmentCandidate(header message.Header, data []byte, readErr error) (attachmentCandidate, bool) {
	// Like the Python client, a part with no Content-Disposition is never an
	// attachment. Those are typically images an HTML body embeds by cid.
	if len(header.Get("Content-Disposition")) == 0 {
		return attachmentCandidate{}, false
	}

	filename := utils.SanitizeFileName(partFilename(header))
	if len(filename) == 0 {
		return attachmentCandidate{}, false
	}

	disposition, _ := contentDispositionOf(header)
	declaredType, _ := contentTypeOf(header)
	sniffedType := utils.GetMimeType(data).String()

	candidate := attachmentCandidate{
		attachment: Attachment{Filename: filename, FileType: sniffedType, Data: data},
		inline:     disposition == "inline",
		contentId:  strings.Trim(strings.TrimSpace(header.Get("Content-Id")), "<>"),
	}

	switch {
	case readErr != nil:
		if !attachmentMimeTypes.MatchString(declaredType) && !attachmentMimeTypes.MatchString(sniffedType) {
			return attachmentCandidate{}, false
		}
		candidate.rejection = "it could not be decoded: " + readErr.Error()
	case attachmentMimeTypes.MatchString(sniffedType):
	case attachmentMimeTypes.MatchString(declaredType):
		candidate.rejection = fmt.Sprintf("its contents are not a supported image or PDF (they look like %s)", sniffedType)
	default:
		return attachmentCandidate{}, false
	}

	return candidate, true
}

// keepReceiptAttachments drops inline images that are HTML decoration: embedded
// by the body (cid:) and small, whether usable or not. Everything else is kept,
// in message order, split into usable attachments and rejected ones.
func keepReceiptAttachments(candidates []attachmentCandidate, bodyHtml string) ([]Attachment, []RejectedAttachment) {
	lowerHtml := strings.ToLower(bodyHtml)
	attachments := make([]Attachment, 0, len(candidates))
	rejected := make([]RejectedAttachment, 0)

	for _, candidate := range candidates {
		embedded := candidate.inline &&
			len(candidate.contentId) > 0 &&
			strings.Contains(lowerHtml, "cid:"+strings.ToLower(candidate.contentId))

		if embedded && len(candidate.attachment.Data) < InlineImageMaxBytes {
			continue
		}

		if len(candidate.rejection) > 0 {
			rejected = append(rejected, RejectedAttachment{Filename: candidate.attachment.Filename, Reason: candidate.rejection})
			continue
		}

		attachments = append(attachments, candidate.attachment)
	}

	return attachments, rejected
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

	contentType, _ := contentTypeOf(header)
	switch contentType {
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
	_, params := contentDispositionOf(header)
	if filename, ok := params["filename"]; ok && len(filename) > 0 {
		return filename
	}

	_, params = contentTypeOf(header)
	return params["name"]
}

// contentTypeOf returns the lowercased media type and parameters of a part,
// text/plain when it declares none.
func contentTypeOf(header message.Header) (string, map[string]string) {
	contentType, params, err := header.ContentType()
	if err != nil {
		return lenientParams(header.Get("Content-Type"))
	}
	return strings.ToLower(contentType), params
}

// contentDispositionOf returns the lowercased disposition and parameters of a part.
func contentDispositionOf(header message.Header) (string, map[string]string) {
	disposition, params, err := header.ContentDisposition()
	if err != nil {
		return lenientParams(header.Get("Content-Disposition"))
	}
	return strings.ToLower(disposition), params
}

// parameterWordDecoder decodes RFC 2047 words in a parameter value, with the
// charsets the charset package registers.
var parameterWordDecoder = &mime.WordDecoder{
	CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
		return message.CharsetReader(charset, input)
	},
}

// lenientParams reads a header that mime.ParseMediaType refuses. That refuses the
// whole header over one malformed parameter, and go-message then returns no
// parameters at all; an unquoted filename with a space in it is enough
// ("filename=Receipt 2024.pdf"). Python's parser read such headers anyway, so
// this splits on semicolons outside quotes and takes each value as written.
// RFC 2231 forms (filename*=) are not decoded here: they are rare in a header
// that is malformed anyway.
func lenientParams(raw string) (string, map[string]string) {
	segments := splitOutsideQuotes(raw, ';')
	params := make(map[string]string)

	for _, segment := range segments[1:] {
		key, value, found := strings.Cut(segment, "=")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if unquoted, ok := unquote(value); ok {
			value = unquoted
		}
		if decoded, err := parameterWordDecoder.DecodeHeader(value); err == nil {
			value = decoded
		}
		params[key] = value
	}

	return strings.ToLower(strings.TrimSpace(segments[0])), params
}

// splitOutsideQuotes splits s on sep, ignoring separators inside a quoted string.
func splitOutsideQuotes(s string, sep rune) []string {
	var segments []string
	var current strings.Builder
	quoted, escaped := false, false

	for _, r := range s {
		switch {
		case escaped:
			escaped = false
		case quoted && r == '\\':
			escaped = true
		case r == '"':
			quoted = !quoted
		case r == sep && !quoted:
			segments = append(segments, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(r)
	}

	return append(segments, current.String())
}

// unquote removes the quotes around a quoted-string and its backslash escapes.
func unquote(value string) (string, bool) {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return value, false
	}

	var builder strings.Builder
	escaped := false
	for _, r := range value[1 : len(value)-1] {
		if !escaped && r == '\\' {
			escaped = true
			continue
		}
		escaped = false
		builder.WriteRune(r)
	}
	return builder.String(), true
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
	// A fresh slice: appending onto the trimmed one would write into the
	// caller's array whenever it has spare capacity.
	block := slices.Concat(bytes.TrimRight(rawHeader, "\r\n"), []byte("\r\n\r\n"))
	entity, err := message.Read(bytes.NewReader(block))
	if entity == nil {
		return gomail.Header{}, err
	}
	return gomail.Header{Header: entity.Header}, nil
}
