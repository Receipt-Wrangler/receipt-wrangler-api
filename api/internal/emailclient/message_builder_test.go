package emailclient

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

// Fixture bytes whose type the mimetype sniffer recognizes from the prefix.
var (
	jpegBytes = append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("JFIF receipt photo")...)
	pngBytes  = append([]byte("\x89PNG\r\n\x1a\n"), []byte("receipt png")...)
	pdfBytes  = []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\ntrailer\n<<>>\n%%EOF\n")
	// heifBytes starts with an ftyp box whose major brand is mif1, which is how
	// a HEIF that is not branded HEIC begins.
	heifBytes = append([]byte("\x00\x00\x00\x18ftypmif1\x00\x00\x00\x00mif1heic"), make([]byte, 32)...)
)

// largeJpeg returns a JPEG-sniffing payload of at least size bytes.
func largeJpeg(size int) []byte {
	return append(append([]byte{}, jpegBytes...), bytes.Repeat([]byte{0x42}, size)...)
}

// mimePart is one node of a test message. A part with Children is a multipart
// container and ContentType must be multipart/<subtype>.
type mimePart struct {
	ContentType string
	Disposition string
	ContentID   string
	// Encoding is the Content-Transfer-Encoding: "base64" encodes Body,
	// anything else writes Body as-is.
	Encoding string
	Body     []byte
	Children []mimePart
}

// buildMessage renders a raw RFC 5322 message from header lines and a root part.
func buildMessage(headers []string, root mimePart) []byte {
	var buffer bytes.Buffer
	for _, header := range headers {
		buffer.WriteString(header)
		buffer.WriteString("\r\n")
	}
	buffer.WriteString("MIME-Version: 1.0\r\n")
	boundaryCounter := 0
	writePart(&buffer, root, &boundaryCounter)
	return buffer.Bytes()
}

func writePart(buffer *bytes.Buffer, part mimePart, boundaryCounter *int) {
	if len(part.Children) > 0 {
		*boundaryCounter++
		boundary := fmt.Sprintf("boundary-%d", *boundaryCounter)
		fmt.Fprintf(buffer, "Content-Type: %s; boundary=\"%s\"\r\n\r\n", part.ContentType, boundary)
		for _, child := range part.Children {
			fmt.Fprintf(buffer, "--%s\r\n", boundary)
			writePart(buffer, child, boundaryCounter)
			// This CRLF belongs to the delimiter below, not to the part body.
			buffer.WriteString("\r\n")
		}
		fmt.Fprintf(buffer, "--%s--", boundary)
		return
	}

	fmt.Fprintf(buffer, "Content-Type: %s\r\n", part.ContentType)
	if len(part.Disposition) > 0 {
		fmt.Fprintf(buffer, "Content-Disposition: %s\r\n", part.Disposition)
	}
	if len(part.ContentID) > 0 {
		fmt.Fprintf(buffer, "Content-ID: <%s>\r\n", part.ContentID)
	}
	if part.Encoding == "base64" {
		buffer.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		encoded := base64.StdEncoding.EncodeToString(part.Body)
		for len(encoded) > 76 {
			buffer.WriteString(encoded[:76] + "\r\n")
			encoded = encoded[76:]
		}
		buffer.WriteString(encoded)
		return
	}
	if len(part.Encoding) > 0 {
		fmt.Fprintf(buffer, "Content-Transfer-Encoding: %s\r\n", part.Encoding)
	}
	buffer.WriteString("\r\n")
	buffer.Write(part.Body)
}

func standardHeaders(extra ...string) []string {
	return append([]string{
		"From: Sender <sender@example.com>",
		"To: Recipient <recipient@example.com>",
		"Subject: Test Receipt",
		"Date: Wed, 20 Oct 2021 10:30:00 +0000",
	}, extra...)
}

// headersWithout returns standardHeaders minus any line starting with prefix,
// followed by the replacement lines in extra.
func headersWithout(prefix string, extra ...string) []string {
	result := make([]string, 0)
	for _, header := range standardHeaders() {
		if !strings.HasPrefix(header, prefix) {
			result = append(result, header)
		}
	}
	return append(result, extra...)
}

func textPart(contentType string, body string) mimePart {
	return mimePart{ContentType: contentType, Body: []byte(body)}
}

func jpegAttachment(filename string) mimePart {
	return mimePart{
		ContentType: "image/jpeg",
		Disposition: fmt.Sprintf("attachment; filename=%q", filename),
		Encoding:    "base64",
		Body:        jpegBytes,
	}
}

func mixed(children ...mimePart) mimePart {
	return mimePart{ContentType: "multipart/mixed", Children: children}
}
