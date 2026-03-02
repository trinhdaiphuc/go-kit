package mailbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// encodeJSONBody JSON-encodes v and sets it as the body of req.
// This is a small internal helper to avoid manual byte-buffer management.
func encodeJSONBody(req *http.Request, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("mailbox: marshal request body: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(b))
	req.ContentLength = int64(len(b))
	return nil
}
