package url

import (
	"net/url"
	"strings"

	"github.com/samber/lo"
)

var validURLExtension = []string{".png", ".jpg", ".jpeg", ".gif", ".svg"}

func PathEscape(fullPath string) string {
	if parse, err := url.Parse(fullPath); err == nil {
		return parse.String()
	}
	return fullPath
}

func ValidImageURL(imageURL string) bool {
	if _, err := url.ParseRequestURI(imageURL); err != nil {
		return false
	}

	return lo.ContainsBy(validURLExtension, func(ext string) bool {
		return strings.Contains(imageURL, ext)
	})
}
