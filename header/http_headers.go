package header

import "encoding/base64"

const (
	HTTPContentType    = "Content-Type"
	HTTPContentLength  = "Content-Length"
	HTTPUserAgent      = "User-Agent"
	HTTPAccept         = "Accept"
	HTTPAcceptLanguage = "Accept-Language"
	HTTPCookie         = "Cookie"
	HTTPCacheControl   = "Cache-Control"
	HTTPAuthorization  = "Authorization"
)

const (
	ContentTypeApplicationJSON   = "application/json"
	ContentTypeApplicationXML    = "application/xml"
	ContentTypeTextPlain         = "text/plain"
	ContentTypeTextHTML          = "text/html"
	ContentTypeFormURLEncoded    = "application/x-www-form-urlencoded"
	ContentTypeMultipartFormData = "multipart/form-data"
)

// BasicAuth See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func BearerAuth(token string) string {
	return "Bearer " + token
}

type Headers map[string]string
