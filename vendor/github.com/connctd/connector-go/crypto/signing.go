package crypto

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"net/http"
)

// SignedHeaderKey defines a header field name
type signedHeaderKey string

const (
	// signedHeaderKeyDate stands for the date header
	signedHeaderKeyDate signedHeaderKey = "Date"
)

// signedHeaderKeys defines a list of headers that are used to build
// the payload-to-be-signed. If a request does not contain all of these
// headers it can't be signed nor validated and thus is invalid.
// The order of keys inside this list defines how the payload-to-be-signed
// is constructed
var signedHeaderKeys = []signedHeaderKey{
	signedHeaderKeyDate,
}

const (
	// SignatureHeaderKey defines header carrying the signature
	SignatureHeaderKey = "Signature"

	// signatureFragmentDelimiter defines how different fragments like headers and
	// body are concatenated to a signable payload. CRLF is already used as a
	// separator for http 1.1 headers and payloads (https://tools.ietf.org/html/rfc7230#page-19)
	// which means underlying libraries should already be aware of correct
	// CRLF handling (e.g. prevent CRLF injection)
	signatureFragmentDelimiter = "\r\n"

	// separates keys from values in constructed payload
	keyValueSeparator = ":"
)

// SignablePayload builds the payload which can be signed
// Method\r\nHost\r\nRequestURI\r\nDate Header Value\r\nBody
// Example: (method):-method-\r\n(url):-scheme-://-host--requestURI-\r\n(Date):Wed, 07 Oct 2020 10:00:00 GMT\r\n(body):{\"hello\":\"world\"}
func SignablePayload(method string, scheme string, host string, requestURI string, headers http.Header, body []byte) ([]byte, error) {
	var b bytes.Buffer

	// write method
	b.WriteString("(method)")
	b.WriteString(keyValueSeparator)
	b.WriteString(method)
	b.WriteString(signatureFragmentDelimiter)

	// write url
	b.WriteString("(url)")
	b.WriteString(keyValueSeparator)
	b.WriteString(scheme + "://")
	b.WriteString(host)
	b.WriteString(requestURI)
	b.WriteString(signatureFragmentDelimiter)

	// write all required headers
	for _, currHeader := range signedHeaderKeys {
		value := headers.Get(string(currHeader))
		if value == "" {
			return []byte{}, ErrorMissingHeader
		}

		b.WriteString("(" + string(currHeader) + ")")
		b.WriteString(keyValueSeparator)
		b.WriteString(value)
		b.WriteString(signatureFragmentDelimiter)
	}

	// write body
	b.WriteString("(body)")
	b.WriteString(keyValueSeparator)
	b.Write(body)

	return b.Bytes(), nil
}

// Verify reports whether signature is a valid signature of message by publicKey. It
// will panic if len(publicKey) is not PublicKeySize
func Verify(publicKey ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// Sign signs the message with privateKey and returns a signature. It will
// panic if len(privateKey) is not PrivateKeySize
func Sign(privateKey ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

// Definition of error cases
var (
	// ErrorMissingHeader is returned when a header used in the canonical request representation is missing
	ErrorMissingHeader = errors.New("signable payload can not be generated since a relevant header is missing")
)
