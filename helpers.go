package core

import (
	"crypto"
	_ "crypto/sha1" // used for SHA1 function
	"encoding/base64"
	"html"
	"html/template"
)

func SHA1(input string) []byte {
	h := crypto.SHA1.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func Base64Encode(bv []byte) string {
	return base64.URLEncoding.EncodeToString(bv)
}

func Clean(val string) string {
	return html.EscapeString(template.JSEscapeString(template.HTMLEscapeString(val)))
}
