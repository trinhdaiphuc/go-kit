package mailbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// filterUnread is the OData filter expression for unread messages.
const filterUnread = "isRead eq false"

// EmailAddress wraps the nested emailAddress object returned by Graph API.
type EmailAddress struct {
	EmailAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"emailAddress"`
}

// Recipient wraps the nested emailAddress object for message recipients.
type Recipient struct {
	EmailAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"emailAddress"`
}

// Body holds the content type and raw content of a message body.
type Body struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// Message represents an email message returned by the Microsoft Graph API.
type Message struct {
	ID                string       `json:"id"`
	Subject           string       `json:"subject"`
	BodyPreview       string       `json:"bodyPreview"`
	ReceivedDateTime  time.Time    `json:"receivedDateTime"`
	IsRead            bool         `json:"isRead"`
	HasAttachments    bool         `json:"hasAttachments"`
	From              EmailAddress `json:"from"`
	ToRecipients      []Recipient  `json:"toRecipients"`
	Body              Body         `json:"body"`
	ConversationID    string       `json:"conversationId"`
	InternetMessageID string       `json:"internetMessageId"`
}

// MessagesResponse is the paged collection returned by Graph API for messages.
type MessagesResponse struct {
	Value    []Message `json:"value"`
	NextLink string    `json:"@odata.nextLink"`
}

// GetMessages retrieves up to top messages from the authenticated user's
// mailbox, ordered by received time descending. An optional OData filter
// expression can be provided via filter (e.g. "isRead eq false").
func (c *MailboxClient) GetMessages(ctx context.Context, top int, filter string) ([]Message, error) {
	endpoint := messageURL(top, filter)

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("mailbox: create messages request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailbox: execute messages request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read
		return nil, fmt.Errorf("mailbox: messages API error (status %d): %s", resp.StatusCode, body)
	}

	var result MessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("mailbox: decode messages response: %w", err)
	}

	return result.Value, nil
}

// GetUnreadMessages is a convenience wrapper that returns up to top unread
// messages from the mailbox.
func (c *MailboxClient) GetUnreadMessages(ctx context.Context, top int) ([]Message, error) {
	return c.GetMessages(ctx, top, filterUnread)
}

// GetMessage retrieves a single message by its Graph API message ID.
func (c *MailboxClient) GetMessage(ctx context.Context, messageID string) (*Message, error) {
	endpoint := messageByIDURL(messageID)

	req, err := c.newRequest(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, fmt.Errorf("mailbox: create message request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailbox: execute message request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read
		return nil, fmt.Errorf("mailbox: message API error (status %d): %s", resp.StatusCode, body)
	}

	var msg Message
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, fmt.Errorf("mailbox: decode message response: %w", err)
	}

	return &msg, nil
}

// MarkAsRead marks the message identified by messageID as read.
func (c *MailboxClient) MarkAsRead(ctx context.Context, messageID string) error {
	endpoint := messageByIDURL(messageID)

	req, err := c.newRequest(ctx, http.MethodPatch, endpoint)
	if err != nil {
		return fmt.Errorf("mailbox: create mark-as-read request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Inline the small JSON body without allocating a byte buffer.
	type patchBody struct {
		IsRead bool `json:"isRead"`
	}
	if err := encodeJSONBody(req, patchBody{IsRead: true}); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailbox: execute mark-as-read request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) //nolint:errcheck // best-effort error body read
		return fmt.Errorf("mailbox: mark-as-read API error (status %d): %s", resp.StatusCode, body)
	}

	return nil
}

// WatchMailbox polls the mailbox every interval and calls handler for each
// new message that was not present during the initial fetch. Runs until ctx
// is cancelled, at which point it returns ctx.Err().
func (c *MailboxClient) WatchMailbox(ctx context.Context, interval time.Duration, handler func(Message)) error {
	log.Printf("mailbox: starting watcher (polling every %v)", interval)

	seen := make(map[string]bool)

	initial, err := c.GetMessages(ctx, 50, "")
	if err != nil {
		return fmt.Errorf("mailbox: initial fetch: %w", err)
	}
	for _, m := range initial {
		seen[m.ID] = true
	}
	log.Printf("mailbox: initialized with %d existing messages", len(seen))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("mailbox: watcher stopped")
			return ctx.Err()
		case <-ticker.C:
			messages, err := c.GetMessages(ctx, 20, "")
			if err != nil {
				log.Printf("mailbox: error fetching messages: %v", err)
				continue
			}
			for _, msg := range messages {
				if !seen[msg.ID] {
					seen[msg.ID] = true
					handler(msg)
				}
			}
		}
	}
}
