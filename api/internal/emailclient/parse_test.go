package emailclient

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func mustParse(t *testing.T, raw []byte) Message {
	t.Helper()
	message, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage() error = %v", err)
	}
	return message
}

// --- Parity with the Python client's test suite -----------------------------

func TestParseMessage_PlainTextBody(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/plain; charset=utf-8", "This is a plain text receipt.")))

	if message.Body != "This is a plain text receipt." {
		t.Errorf("Body = %q", message.Body)
	}
	if message.BodyHtml != "" {
		t.Errorf("BodyHtml = %q, want empty", message.BodyHtml)
	}
}

func TestParseMessage_HtmlOnlyBody(t *testing.T) {
	htmlBody := "<p>Your order total is <b>$45.00</b></p>"
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/html; charset=utf-8", htmlBody)))

	if message.Body != "Your order total is $45.00" {
		t.Errorf("Body = %q", message.Body)
	}
	if message.BodyHtml != htmlBody {
		t.Errorf("BodyHtml = %q, want %q", message.BodyHtml, htmlBody)
	}
}

func TestParseMessage_AlternativePrefersPlainTextAndKeepsHtml(t *testing.T) {
	htmlBody := "<p>HTML version</p>"
	message := mustParse(t, buildMessage(standardHeaders(), mimePart{
		ContentType: "multipart/alternative",
		Children: []mimePart{
			textPart("text/plain; charset=utf-8", "Plain text version"),
			textPart("text/html; charset=utf-8", htmlBody),
		},
	}))

	if message.Body != "Plain text version" {
		t.Errorf("Body = %q", message.Body)
	}
	if message.BodyHtml != htmlBody {
		t.Errorf("BodyHtml = %q", message.BodyHtml)
	}
}

func TestParseMessage_NoBody(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mixed(jpegAttachment("receipt.jpg"))))

	if message.Body != "" || message.BodyHtml != "" {
		t.Errorf("Body = %q, BodyHtml = %q, want both empty", message.Body, message.BodyHtml)
	}
	if len(message.Attachments) != 1 {
		t.Fatalf("Attachments = %d, want 1", len(message.Attachments))
	}
}

func TestParseMessage_TextAttachmentIsNeitherBodyNorAttachment(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mixed(
		textPart("text/plain; charset=utf-8", "Receipt body"),
		mimePart{
			ContentType: "text/plain; charset=utf-8",
			Disposition: `attachment; filename="notes.txt"`,
			Body:        []byte("Attached text content"),
		},
	)))

	if message.Body != "Receipt body" {
		t.Errorf("Body = %q", message.Body)
	}
	if message.BodyHtml != "" {
		t.Errorf("BodyHtml = %q", message.BodyHtml)
	}
	if len(message.Attachments) != 0 {
		t.Errorf("Attachments = %+v, want none", message.Attachments)
	}
}

func TestParseMessage_CollapsesWhitespace(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/plain; charset=utf-8", "  Line one  \n\n\n\n  Line two  ")))

	// Spaces collapse to one, three or more newlines to two, then trimmed.
	if message.Body != "Line one \n\n Line two" {
		t.Errorf("Body = %q", message.Body)
	}
}

func TestParseMessage_CollapsesWhitespaceAcrossCrlf(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/plain; charset=utf-8", "Line one\r\n\r\n\r\n\r\nLine two")))

	if message.Body != "Line one\n\nLine two" {
		t.Errorf("Body = %q", message.Body)
	}
}

func TestParseMessage_MultipartHtmlFallback(t *testing.T) {
	htmlBody := "<h1>Receipt</h1><p>Total: $10</p>"
	message := mustParse(t, buildMessage(standardHeaders(), mixed(textPart("text/html; charset=utf-8", htmlBody))))

	if message.Body != "ReceiptTotal: $10" {
		t.Errorf("Body = %q", message.Body)
	}
	if message.BodyHtml != htmlBody {
		t.Errorf("BodyHtml = %q", message.BodyHtml)
	}
}

func TestParseMessage_BodyOnlyEmailHasContent(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mixed(textPart("text/plain; charset=utf-8", "Your order total: $25.00"))))

	if !message.HasContent() {
		t.Error("HasContent() = false, want true")
	}
	if message.Body != "Your order total: $25.00" || message.BodyHtml != "" || len(message.Attachments) != 0 {
		t.Errorf("message = %+v", message)
	}
}

func TestParseMessage_NoBodyNoAttachmentsHasNoContent(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mimePart{ContentType: "multipart/mixed", Children: []mimePart{
		{ContentType: "application/octet-stream", Disposition: "attachment", Body: []byte("x")},
	}}))

	if message.HasContent() {
		t.Errorf("HasContent() = true for %+v", message)
	}
}

func TestParseMessage_BodyWithAttachment(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mixed(
		textPart("text/plain; charset=utf-8", "Order confirmation"),
		jpegAttachment("receipt.jpg"),
	)))

	if message.Body != "Order confirmation" || message.BodyHtml != "" {
		t.Errorf("Body = %q, BodyHtml = %q", message.Body, message.BodyHtml)
	}
	if len(message.Attachments) != 1 {
		t.Fatalf("Attachments = %d, want 1", len(message.Attachments))
	}
	attachment := message.Attachments[0]
	if attachment.Filename != "receipt.jpg" || attachment.FileType != "image/jpeg" || !bytes.Equal(attachment.Data, jpegBytes) {
		t.Errorf("attachment = %q %q %d bytes", attachment.Filename, attachment.FileType, len(attachment.Data))
	}
}

func TestParseMessage_FromNameAndAddress(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/plain", "x")))

	// The Python client kept a trailing space on the name ("Sender "); the
	// parsed name is trimmed.
	if message.From.Name != "Sender" || message.From.Email != "sender@example.com" {
		t.Errorf("From = %+v", message.From)
	}
	if message.To != "recipient@example.com" {
		t.Errorf("To = %q", message.To)
	}
	if message.Subject != "Test Receipt" {
		t.Errorf("Subject = %q", message.Subject)
	}
}

func TestParseMessage_Date(t *testing.T) {
	cases := []struct {
		name string
		date string
		want time.Time
	}{
		{"utc offset", "Wed, 20 Oct 2021 10:30:00 +0000", time.Date(2021, 10, 20, 10, 30, 0, 0, time.UTC)},
		{"utc zone comment", "Wed, 20 Oct 2021 10:30:00 +0000 (UTC)", time.Date(2021, 10, 20, 10, 30, 0, 0, time.UTC)},
		// The Python client replaced the zone without converting, so this read
		// as 22:00 UTC on Jan 1. The instant is 03:00 UTC on Jan 2.
		{"negative offset converts to utc", "Mon, 01 Jan 2024 22:00:00 -0500", time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)},
		{"no weekday", "20 Oct 2021 10:30:00 +0200", time.Date(2021, 10, 20, 8, 30, 0, 0, time.UTC)},
		{"unparseable is zero", "sometime last week", time.Time{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers := headersWithout("Date:", "Date: "+tc.date)
			message := mustParse(t, buildMessage(headers, textPart("text/plain", "x")))

			if !message.Date.Equal(tc.want) {
				t.Errorf("Date = %v, want %v", message.Date, tc.want)
			}
			if !message.Date.IsZero() && message.Date.Location() != time.UTC {
				t.Errorf("Date location = %v, want UTC", message.Date.Location())
			}
		})
	}
}

func TestStripHtmlTags(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string
	}{
		{"basic", "<p>Hello <b>World</b></p>", "Hello World"},
		{"complex", "<html><body><h1>Title</h1><p>Content</p></body></html>", "TitleContent"},
		{"empty", "", ""},
		{"plain text unchanged", "no html here", "no html here"},
		{"entities decoded", "<p>Fish &amp; Chips &lt;3</p>", "Fish & Chips <3"},
		{"style dropped", "<style>.total { color: red; }</style><p>Total: $5</p>", "Total: $5"},
		{"script dropped", "<p>Total</p><script>track('open');</script><p>$5</p>", "Total$5"},
		{"head style dropped", "<html><head><style>body{margin:0}</style></head><body>Paid</body></html>", "Paid"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StripHtmlTags(tc.html); got != tc.want {
				t.Errorf("StripHtmlTags(%q) = %q, want %q", tc.html, got, tc.want)
			}
		})
	}
}

func TestParseMessage_StyleNeverReachesBodyText(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), textPart("text/html; charset=utf-8",
		"<html><head><style>td { font-family: Arial; }</style></head><body><td>Total $12.00</td></body></html>")))

	if message.Body != "Total $12.00" {
		t.Errorf("Body = %q", message.Body)
	}
}

// --- Attachments --------------------------------------------------------------

func TestParseMessage_AttachmentFilenamesAreSanitized(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		want     string
	}{
		{"traversal reduced to base", "../../../../evil.jpg", "evil.jpg"},
		{"module overwrite reduced to base", "../imap-client/utils.py", "utils.py"},
		{"legitimate unchanged", "receipt.jpg", "receipt.jpg"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, buildMessage(standardHeaders(), mixed(jpegAttachment(tc.filename))))

			if len(message.Attachments) != 1 || message.Attachments[0].Filename != tc.want {
				t.Errorf("Attachments = %+v, want one named %q", message.Attachments, tc.want)
			}
		})
	}
}

func TestParseMessage_UnusableFilenamesAreSkipped(t *testing.T) {
	cases := []struct {
		name        string
		disposition string
	}{
		{"dotdot", `attachment; filename=".."`},
		{"missing filename", "attachment"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, buildMessage(standardHeaders(), mixed(mimePart{
				ContentType: "image/jpeg", Disposition: tc.disposition, Encoding: "base64", Body: jpegBytes,
			})))

			if len(message.Attachments) != 0 {
				t.Errorf("Attachments = %+v, want none", message.Attachments)
			}
		})
	}
}

func TestParseMessage_AttachmentTypeIsSniffedNotDeclared(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        []byte
		wantType    string
	}{
		{"jpeg accepted", "image/jpeg", jpegBytes, "image/jpeg"},
		{"png accepted", "image/png", pngBytes, "image/png"},
		{"pdf accepted", "application/pdf", pdfBytes, "application/pdf"},
		// The Python client trusted the declared type and skipped these.
		{"pdf declared as octet-stream accepted", "application/octet-stream", pdfBytes, "application/pdf"},
		{"jpeg declared as nonstandard image/jpg accepted", "image/jpg", jpegBytes, "image/jpeg"},
		// ...and accepted this one.
		{"text declared as jpeg rejected", "image/jpeg", []byte("this is not an image"), ""},
		{"plain text rejected", "text/plain", []byte("notes"), ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, buildMessage(standardHeaders(), mixed(mimePart{
				ContentType: tc.contentType, Disposition: `attachment; filename="file.bin"`, Encoding: "base64", Body: tc.body,
			})))

			if tc.wantType == "" {
				if len(message.Attachments) != 0 {
					t.Errorf("Attachments = %+v, want none", message.Attachments)
				}
				return
			}
			if len(message.Attachments) != 1 || message.Attachments[0].FileType != tc.wantType {
				t.Errorf("Attachments = %+v, want one of type %q", message.Attachments, tc.wantType)
			}
		})
	}
}

func TestParseMessage_PartWithoutDispositionIsNotAnAttachment(t *testing.T) {
	// Parity: the Python client only took parts carrying a Content-Disposition.
	message := mustParse(t, buildMessage(standardHeaders(), mixed(
		textPart("text/plain", "body"),
		mimePart{ContentType: `image/jpeg; name="photo.jpg"`, Encoding: "base64", Body: jpegBytes},
	)))

	if len(message.Attachments) != 0 {
		t.Errorf("Attachments = %+v, want none", message.Attachments)
	}
}

func TestParseMessage_DuplicateFilenamesAreKeptSeparately(t *testing.T) {
	second := append(append([]byte{}, jpegBytes...), []byte("second")...)
	message := mustParse(t, buildMessage(standardHeaders(), mixed(
		jpegAttachment("image001.jpg"),
		mimePart{ContentType: "image/jpeg", Disposition: `attachment; filename="image001.jpg"`, Encoding: "base64", Body: second},
	)))

	if len(message.Attachments) != 2 {
		t.Fatalf("Attachments = %d, want 2", len(message.Attachments))
	}
	if !bytes.Equal(message.Attachments[0].Data, jpegBytes) || !bytes.Equal(message.Attachments[1].Data, second) {
		t.Error("attachment bytes were not kept per part")
	}
}

func TestParseMessage_InlineImageRule(t *testing.T) {
	small := largeJpeg(1024)
	large := largeJpeg(InlineImageMaxBytes)

	related := func(html string, image mimePart) []byte {
		return buildMessage(standardHeaders(), mimePart{ContentType: "multipart/related", Children: []mimePart{
			textPart("text/html; charset=utf-8", html),
			image,
		}})
	}

	cases := []struct {
		name    string
		raw     []byte
		wantLen int
	}{
		{
			"small cid-referenced inline image is decoration",
			related(`<p>Thanks</p><img src="cid:logo@example">`, mimePart{
				ContentType: "image/png", Disposition: `inline; filename="image001.png"`, ContentID: "logo@example", Encoding: "base64", Body: append(append([]byte{}, pngBytes...), small...),
			}),
			0,
		},
		{
			"large cid-referenced inline image is kept (iOS Mail photo)",
			related(`<div><img src="cid:photo@example"></div>`, mimePart{
				ContentType: "image/jpeg", Disposition: `inline; filename="IMG_1234.jpeg"`, ContentID: "photo@example", Encoding: "base64", Body: large,
			}),
			1,
		},
		{
			"small inline image not referenced by the html is kept",
			related(`<p>See attached</p>`, mimePart{
				ContentType: "image/jpeg", Disposition: `inline; filename="receipt.jpg"`, ContentID: "unreferenced@example", Encoding: "base64", Body: small,
			}),
			1,
		},
		{
			"small inline image without a content id is kept (Apple Mail)",
			related(`<p>Receipt</p>`, mimePart{
				ContentType: "image/jpeg", Disposition: `inline; filename="receipt.jpg"`, Encoding: "base64", Body: small,
			}),
			1,
		},
		{
			"small attachment referenced by cid is kept",
			related(`<img src="cid:att@example">`, mimePart{
				ContentType: "image/jpeg", Disposition: `attachment; filename="receipt.jpg"`, ContentID: "att@example", Encoding: "base64", Body: small,
			}),
			1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, tc.raw)
			if len(message.Attachments) != tc.wantLen {
				t.Errorf("Attachments = %d, want %d", len(message.Attachments), tc.wantLen)
			}
		})
	}
}

// --- Headers ------------------------------------------------------------------

func TestParseMessage_DecodesEncodedHeaders(t *testing.T) {
	headers := []string{
		"From: =?UTF-8?B?w4lsb2RpZSBNYXJ0aW4=?= <elodie@example.com>",
		"To: recipient@example.com",
		"Subject: =?UTF-8?Q?Re=C3=A7u_n=C2=B0_42?=",
		"Date: Wed, 20 Oct 2021 10:30:00 +0000",
	}
	message := mustParse(t, buildMessage(headers, mixed(
		textPart("text/plain", "x"),
		mimePart{ContentType: "image/jpeg", Disposition: `attachment; filename="=?UTF-8?B?cmXDp3UuanBn?="`, Encoding: "base64", Body: jpegBytes},
		mimePart{ContentType: "image/jpeg", Disposition: `attachment; filename*=UTF-8''facture%20%C3%A9t%C3%A9.jpg`, Encoding: "base64", Body: jpegBytes},
	)))

	if message.From.Name != "Élodie Martin" {
		t.Errorf("From.Name = %q", message.From.Name)
	}
	if message.Subject != "Reçu n° 42" {
		t.Errorf("Subject = %q", message.Subject)
	}
	if len(message.Attachments) != 2 || message.Attachments[0].Filename != "reçu.jpg" || message.Attachments[1].Filename != "facture été.jpg" {
		t.Errorf("Attachments = %+v", message.Attachments)
	}
}

func TestParseMessage_DecodesNonUtf8Charsets(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		body        []byte
		want        string
	}{
		// 0xE9 is é and 0x80 the euro sign in windows-1252.
		{"windows-1252", "text/plain; charset=windows-1252", []byte("Caf\xe9 total \x8012"), "Café total €12"},
		{"iso-8859-1", "text/plain; charset=iso-8859-1", []byte("Caf\xe9 au lait"), "Café au lait"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := mustParse(t, buildMessage(standardHeaders(), mimePart{ContentType: tc.contentType, Body: tc.body}))
			if message.Body != tc.want {
				t.Errorf("Body = %q, want %q", message.Body, tc.want)
			}
		})
	}
}

func TestParseMessage_UnknownCharsetPartIsSkipped(t *testing.T) {
	message := mustParse(t, buildMessage(standardHeaders(), mixed(
		textPart("text/plain; charset=x-no-such-charset", "unreadable"),
		textPart("text/html; charset=utf-8", "<p>Total $3</p>"),
		jpegAttachment("receipt.jpg"),
	)))

	if message.Body != "Total $3" {
		t.Errorf("Body = %q", message.Body)
	}
	if len(message.Attachments) != 1 {
		t.Errorf("Attachments = %d, want 1", len(message.Attachments))
	}
}

func TestParseMessage_MissingToAndDateAreAllowed(t *testing.T) {
	// The Python client crashed on a missing To header, failing the whole poll.
	headers := []string{"From: sender@example.com", "Subject: Receipt"}
	message := mustParse(t, buildMessage(headers, textPart("text/plain", "Total $3")))

	if message.To != "" || !message.Date.IsZero() {
		t.Errorf("To = %q, Date = %v", message.To, message.Date)
	}
	if message.From.Email != "sender@example.com" {
		t.Errorf("From = %+v", message.From)
	}
}

func TestParseMessage_UndisclosedRecipients(t *testing.T) {
	headers := headersWithout("To:", "To: undisclosed-recipients:;")
	message := mustParse(t, buildMessage(headers, textPart("text/plain", "x")))

	if message.To != "" {
		t.Errorf("To = %q, want empty", message.To)
	}
}

func TestParseMessage_SenderMustBeUnambiguous(t *testing.T) {
	cases := []struct {
		name    string
		headers []string
		wantErr error
	}{
		{"missing from", headersWithout("From:"), ErrMissingFrom},
		{"empty from", headersWithout("From:", "From: "), ErrMissingFrom},
		{"two from headers", headersWithout("From:", "From: attacker@evil.example", "From: sender@example.com"), ErrAmbiguousFrom},
		{"two addresses in one from", headersWithout("From:", "From: a@example.com, b@example.com"), ErrAmbiguousFrom},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMessage(buildMessage(tc.headers, textPart("text/plain", "x")))
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestParseMessage_MalformedMultipartIsAnError(t *testing.T) {
	raw := []byte(strings.Join([]string{
		"From: sender@example.com",
		"Subject: broken",
		"MIME-Version: 1.0",
		`Content-Type: multipart/mixed; boundary="b"`,
		"",
		"--b",
		"Content-Type: text/plain",
		"",
		"unterminated part with no closing boundary",
	}, "\r\n"))

	if _, err := ParseMessage(raw); err == nil {
		t.Error("ParseMessage() error = nil, want an error for a truncated multipart")
	}
}

func TestParseHeaderFields(t *testing.T) {
	cases := []struct {
		name        string
		header      string
		wantFrom    []string
		wantSubject string
	}{
		{"single", "From: Bob <Bob@Example.com>\r\nSubject: Receipt 1\r\n", []string{"Bob@Example.com"}, "Receipt 1"},
		{"encoded subject", "From: a@x.example\r\nSubject: =?UTF-8?Q?Re=C3=A7u?=\r\n", []string{"a@x.example"}, "Reçu"},
		{"two from headers", "From: a@x.example\r\nFrom: b@y.example\r\nSubject: s\r\n", []string{"a@x.example", "b@y.example"}, "s"},
		{"unparseable from", "From: not an address\r\nSubject: s\r\n", nil, "s"},
		{"empty", "", nil, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := ParseHeaderFields([]byte(tc.header))
			if strings.Join(fields.From, ",") != strings.Join(tc.wantFrom, ",") || fields.Subject != tc.wantSubject {
				t.Errorf("ParseHeaderFields() = %+v", fields)
			}
		})
	}
}
