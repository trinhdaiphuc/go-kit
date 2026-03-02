package mailbox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Attachment represents an email attachment returned by the Microsoft Graph API.
type Attachment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ContentType  string `json:"contentType"`
	Size         int    `json:"size"`
	IsInline     bool   `json:"isInline"`
	ContentBytes string `json:"contentBytes"` // Base64-encoded content
	ContentID    string `json:"contentId"`
	OdataType    string `json:"@odata.type"`
}

// AttachmentsResponse is the paged collection returned by Graph API for
// message attachments.
type AttachmentsResponse struct {
	Value    []Attachment `json:"value"`
	NextLink string       `json:"@odata.nextLink"`
}

// GetAttachments retrieves all attachments for the message identified by
// messageID.
func (c *MailboxClient) GetAttachments(ctx context.Context, messageID string) ([]Attachment, error) {
	endpoint := attachmentsURL(messageID)

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("mailbox: create attachments request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailbox: execute attachments request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read
		return nil, fmt.Errorf("mailbox: attachments API error (status %d): %s", resp.StatusCode, body)
	}

	var result AttachmentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("mailbox: decode attachments response: %w", err)
	}

	return result.Value, nil
}

// SaveAttachment decodes the base64-encoded ContentBytes of att and writes the
// raw bytes to filePath, creating or truncating the file as necessary.
func SaveAttachment(att Attachment, filePath string) error {
	if att.ContentBytes == "" {
		return fmt.Errorf("mailbox: attachment %q has no content bytes", att.Name)
	}

	decoded, err := base64.StdEncoding.DecodeString(att.ContentBytes)
	if err != nil {
		return fmt.Errorf("mailbox: decode base64 content for %q: %w", att.Name, err)
	}

	if err := os.WriteFile(filePath, decoded, 0o600); err != nil {
		return fmt.Errorf("mailbox: write attachment %q to %q: %w", att.Name, filePath, err)
	}

	return nil
}
