package grpcrequest

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
)

var (
	ErrMetadataNotFound = errors.New("metadata not found")
	LogErrorFn          = func(format string, args ...interface{}) {}
)

func ParseHeader(ctx context.Context, extraKeys ...string) (result Metadata, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		LogErrorFn("failed to parse metadata from incoming request")
		err = ErrMetadataNotFound
		return
	}

	result = parseMetadata(md, extraKeys)
	return
}

func parseMetadata(md metadata.MD, extraKeys []string) (result Metadata) {
	for _, fn := range parseMetadataFns {
		fn(md, &result)
	}

	if len(extraKeys) != 0 {
		parseExtras(extraKeys)(md, &result)
	}

	return
}

var (
	parseMetadataFns = []ParseMetadataFn{
		parseDeviceID,
		parseDeviceOS,
		parsePlatform,
		parseUserAgent,
		parseAppVersion,
		parseAuthorization,
		parseDensity,
		parseCookies,
		parseZLPToken,
		parseAccessToken,
		parseUserID,
		parseAcceptLanguage,
	}
)

type ParseMetadataFn func(md metadata.MD, result *Metadata)

const (
	HeaderCookie         = "cookie"
	HeaderDeviceID       = "x-device-id"
	HeaderDeviceOS       = "x-device-os"
	HeaderPlatform       = "x-platform"
	HeaderUserAgent      = "user-agent"
	HeaderAppVersion     = "x-app-version"
	HeaderAuthorization  = "authorization"
	HeaderDensity        = "x-density"
	HeaderAccessToken    = "x-access-token"
	HeaderUserID         = "x-user-id"
	HeaderAcceptLanguage = "accept-language"
	CookieZLPToken       = "zlp_token"
)

func parseDeviceID(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderDeviceID)
	if len(values) == 0 {
		return
	}

	result.DeviceID = values[0]
}

func parseDeviceOS(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderDeviceOS)
	if len(values) == 0 {
		return
	}

	result.DeviceOS = values[0]
}

func parsePlatform(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderPlatform)
	if len(values) == 0 {
		return
	}

	result.Platform = values[0]
}

func parseUserAgent(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderUserAgent)
	if len(values) == 0 {
		return
	}

	result.UserAgent = values[0]
}

func parseAppVersion(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderAppVersion)
	if len(values) == 0 {
		return
	}

	result.AppVersion = values[0]
}

const (
	authorizationPair = 2
)

// Authorization format: Bearer XYZ
func parseAuthorization(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderAuthorization)
	if len(values) == 0 {
		return
	}

	arr := strings.Split(values[0], " ")
	if len(arr) != authorizationPair {
		LogErrorFn("wrong authorization format")
		return
	}

	result.Authorization = arr[1]
}

func parseDensity(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderDensity)
	if len(values) == 0 {
		return
	}

	result.Density = values[0]
}

const (
	cookieKVPair        = 2
	defaultCookieArrLen = 12
)

func parseCookies(md metadata.MD, result *Metadata) {
	result.Cookies = make(map[string][]string)
	values := md.Get(HeaderCookie)
	if len(values) == 0 {
		return
	}

	cookiesValues := strings.Split(strings.TrimSpace(values[0]), ";")

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

func parseExtras(extraKeys []string) ParseMetadataFn {
	return func(md metadata.MD, result *Metadata) {
		result.Extras = make(map[string][]string, len(extraKeys))
		for _, k := range extraKeys {
			result.Extras[k] = md.Get(k)
		}
	}
}

func parseZLPToken(_ metadata.MD, result *Metadata) {
	values, ok := result.Cookies[CookieZLPToken]
	if !ok || len(values) == 0 {
		return
	}

	result.ZLPToken = values[0]
}

func parseAccessToken(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderAccessToken)
	if len(values) == 0 {
		return
	}

	result.AccessToken = values[0]
}

func parseUserID(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderUserID)
	if len(values) == 0 {
		return
	}

	result.UserID = values[0]
}

func parseAcceptLanguage(md metadata.MD, result *Metadata) {
	values := md.Get(HeaderAcceptLanguage)
	if len(values) == 0 {
		return
	}

	result.AcceptLanguage = values[0]
}
