package header

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/trinhdaiphuc/go-kit/log"
)

type Headers map[string]string

type parser struct {
	metadataFns []ParseMetadataFn
	extraKeys   []string
}

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=headermock
type Parser interface {
	ParseHeader(headers http.Header) (result *Metadata, err error)
}

func NewParser(opts ...Option) Parser {
	p := defaultParser
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *parser) ParseHeader(header http.Header) (result *Metadata, err error) {
	result = p.parseMetadata(header)
	return
}

func (p *parser) parseMetadata(header http.Header) (result *Metadata) {
	result = &Metadata{}
	for _, fn := range p.metadataFns {
		fn(header, result)
	}

	if len(p.extraKeys) != 0 {
		ParseExtras(p.extraKeys)(header, result)
	}

	return
}

var (
	parseMetadataFns = []ParseMetadataFn{
		ParseDeviceID,
		ParseDeviceOS,
		ParsePlatform,
		ParseUserAgent,
		ParseAppVersion,
		ParseAuthorization,
		ParseDensity,
		ParseCookies,
		ParseZLPToken,
		ParseAccessToken,
		ParseUserID,
		ParseAcceptLanguage,
		ParseRealIP,
	}
	defaultParser = &parser{
		metadataFns: parseMetadataFns,
		extraKeys:   nil,
	}
)

type ParseMetadataFn func(header http.Header, result *Metadata)

const (
	HeaderCookie               = "cookie"
	HeaderDeviceID             = "x-device-id"
	HeaderDeviceOS             = "x-device-os"
	HeaderPlatform             = "x-platform"
	HeaderUserAgent            = "user-agent"
	HeaderAppVersion           = "x-app-version"
	HeaderAuthorization        = "authorization"
	HeaderDensity              = "x-density"
	HeaderAccessToken          = "x-access-token"
	HeaderUserID               = "x-user-id"
	HeaderAcceptLanguage       = "accept-language"
	HeaderRealIP               = "x-real-ip"
	HeaderXForwardedFor        = "x-forwarded-for"
	HeaderXAuthenticatedUserID = "X-Authenticated-Userid"
	CookieZLPToken             = "zlp_token"
)

func ParseDeviceID(header http.Header, result *Metadata) {
	values := header.Get(HeaderDeviceID)
	if len(values) == 0 {
		return
	}

	result.DeviceID = values
}

func ParseDeviceOS(header http.Header, result *Metadata) {
	values := header.Get(HeaderDeviceOS)
	if len(values) == 0 {
		return
	}

	result.DeviceOS = values
}

func ParsePlatform(header http.Header, result *Metadata) {
	values := header.Get(HeaderPlatform)
	if len(values) == 0 {
		return
	}

	result.Platform = values
}

func ParseUserAgent(header http.Header, result *Metadata) {
	values := header.Get(HeaderUserAgent)
	if len(values) == 0 {
		return
	}

	result.UserAgent = values
}

func ParseAppVersion(header http.Header, result *Metadata) {
	values := header.Get(HeaderAppVersion)
	if len(values) == 0 {
		return
	}

	result.AppVersion = values
}

const (
	authorizationPair = 2
)

// ParseAuthorization format: Bearer XYZ
func ParseAuthorization(header http.Header, result *Metadata) {
	values := header.Get(HeaderAuthorization)
	if len(values) == 0 {
		return
	}

	arr := strings.Split(values, " ")
	if len(arr) != authorizationPair {
		log.Bg().Error("wrong authorization format")
		return
	}

	result.Authorization = arr[1]
}

func ParseDensity(header http.Header, result *Metadata) {
	values := header.Get(HeaderDensity)
	if len(values) == 0 {
		return
	}

	result.Density = values
}

const (
	cookieKVPair        = 2
	defaultCookieArrLen = 12
)

func ParseCookies(header http.Header, result *Metadata) {
	result.Cookies = make(map[string][]string)
	values := header.Get(HeaderCookie)
	if len(values) == 0 {
		return
	}

	cookiesValues := strings.Split(strings.TrimSpace(values), ";")

	result.Cookies = make(map[string][]string, len(cookiesValues))
	for _, v := range cookiesValues {
		value := strings.TrimSpace(v)
		if len(value) == 0 {
			continue
		}
		idx := strings.Index(value, "=")
		if idx == 0 {
			continue
		}
		if idx < 0 {
			strArr := make([]string, 0, defaultCookieArrLen)
			strArr = append(strArr, "")
			result.Cookies[value] = strArr
			continue
		}
		kv, val := value[:idx], value[idx+1:]
		if len(kv) == 0 {
			continue
		}
		strArr, ok := result.Cookies[kv]
		if !ok {
			strArr = make([]string, 0, defaultCookieArrLen)
		}
		strArr = append(strArr, val)
		result.Cookies[kv] = strArr
	}
}

func ParseExtras(extraKeys []string) ParseMetadataFn {
	return func(header http.Header, result *Metadata) {
		result.Extras = make(map[string][]string, len(extraKeys))
		for _, k := range extraKeys {
			result.Extras[k] = header.Values(k)
		}
	}
}

func ParseZLPToken(_ http.Header, result *Metadata) {
	values, ok := result.Cookies[CookieZLPToken]
	if !ok || len(values) == 0 {
		return
	}

	result.ZLPToken = values[0]
}

func ParseAccessToken(header http.Header, result *Metadata) {
	values := header.Get(HeaderAccessToken)
	if len(values) == 0 {
		return
	}

	result.AccessToken = values
}

func ParseUserID(header http.Header, result *Metadata) {
	values := header.Get(HeaderUserID)
	if len(values) == 0 {
		return
	}

	result.UserID = values
}

func ParseAcceptLanguage(header http.Header, result *Metadata) {
	values := header.Get(HeaderAcceptLanguage)
	if len(values) == 0 {
		return
	}

	result.AcceptLanguage = values
}

func ParseRealIP(header http.Header, result *Metadata) {
	realIP := header.Get(HeaderRealIP)
	// Check list of IP in X-Forwarded-For and return the first global address
	for _, address := range strings.Split(header.Get(HeaderXForwardedFor), ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := isPrivateAddress(address)
		if !isPrivate && err == nil {
			result.UserIP = address
			return
		}
	}

	// If nothing succeed, return X-Real-IP
	result.UserIP = realIP
}

func ParseXAuthenticatedUserID(header http.Header, result *Metadata) {
	values := header.Get(HeaderXAuthenticatedUserID)
	if len(values) == 0 {
		return
	}

	result.UserID = values
}

func isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	if ipAddress.IsLoopback() || ipAddress.IsPrivate() {
		return true, nil
	}

	return false, nil
}
