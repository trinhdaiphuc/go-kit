package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidImageURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Valid PNG URL",
			url:      "https://example.com/image.png",
			expected: true,
		},
		{
			name:     "Valid JPG URL",
			url:      "https://example.com/image.jpg",
			expected: true,
		},
		{
			name:     "Valid JPEG URL",
			url:      "https://example.com/image.jpeg",
			expected: true,
		},
		{
			name:     "Valid GIF URL",
			url:      "https://example.com/image.gif",
			expected: true,
		},
		{
			name:     "Valid SVG URL",
			url:      "https://example.com/image.svg",
			expected: true,
		},
		{
			name:     "URL with query string",
			url:      "https://example.com/image.png?key=value",
			expected: true,
		},
		{
			name:     "URL with fragment",
			url:      "https://example.com/image.png#fragment",
			expected: true,
		},
		{
			name:     "Invalid extension",
			url:      "https://example.com/image.txt",
			expected: false,
		},
		{
			name:     "No extension",
			url:      "https://example.com/image",
			expected: false,
		},
		{
			name:     "Invalid URL",
			url:      "not a url",
			expected: false,
		},
		{
			name:     "Empty string",
			url:      "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidImageURL(tc.url); got != tc.expected {
				assert.Equalf(t, tc.expected, got, "ValidImageURL() = %v, want %v", got, tc.expected)
			}
		})
	}
}
