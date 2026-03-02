package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/trinhdaiphuc/go-kit/mailbox"
)

func loadConfig() (*mailbox.Config, error) {
	// Load .env file if it exists (ignored when not present).
	_ = godotenv.Load()

	cfg := &mailbox.Config{
		Username:     os.Getenv("MICROSOFT_USERNAME"),
		Password:     os.Getenv("MICROSOFT_PASSWORD"),
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		TenantID:     os.Getenv("AZURE_TENANT_ID"),
	}

	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("MICROSOFT_USERNAME and MICROSOFT_PASSWORD are required")
	}
	if cfg.ClientID == "" || cfg.TenantID == "" {
		return nil, fmt.Errorf("AZURE_CLIENT_ID and AZURE_TENANT_ID are required")
	}
	if cfg.ClientID == "your_client_id_here" || cfg.TenantID == "your_tenant_id_here" {
		return nil, fmt.Errorf("please configure AZURE_CLIENT_ID and AZURE_TENANT_ID in .env file")
	}

	return cfg, nil
}

func printMessage(msg mailbox.Message) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("📧 Subject: %s\n", msg.Subject)
	fmt.Printf("📬 From: %s <%s>\n", msg.From.EmailAddress.Name, msg.From.EmailAddress.Address)
	fmt.Printf("📅 Received: %s\n", msg.ReceivedDateTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("📖 Read: %v | 📎 Attachments: %v\n", msg.IsRead, msg.HasAttachments)
	fmt.Println("───────────────────────────────────────────────────────────────")

	preview := msg.BodyPreview
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Printf("Preview: %s\n", preview)
	fmt.Println()
}

func main() {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	client, err := mailbox.NewMailboxClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create mailbox client: %v", err)
	}

	// Fetch recent messages.
	fmt.Println("\n📥 Fetching recent messages...")
	messages, err := client.GetMessages(ctx, 10, "")
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("\nFound %d messages:\n\n", len(messages))
	for _, msg := range messages {
		printMessage(msg)
	}

	// Read the first message and its attachments.
	if len(messages) > 0 {
		firstMsg := messages[0]
		fmt.Println("\n📄 Reading first message details...")
		fmt.Printf("Message ID: %s\n", firstMsg.ID)
		fmt.Printf("Subject: %s\n", firstMsg.Subject)

		fullMessage, err := client.GetMessage(ctx, firstMsg.ID)
		if err != nil {
			log.Printf("Failed to get message details: %v", err)
		} else {
			fmt.Println("\n📋 Full Message Body:")
			fmt.Println("───────────────────────────────────────────────────────────────")
			fmt.Printf("Content-Type: %s\n", fullMessage.Body.ContentType)
			content := fullMessage.Body.Content
			if len(content) > 500 {
				content = content[:500] + "\n... (truncated)"
			}
			fmt.Printf("Content:\n%s\n", content)
		}

		if firstMsg.HasAttachments {
			fmt.Println("\n📎 Fetching attachments for first message...")
			attachments, err := client.GetAttachments(ctx, firstMsg.ID)
			if err != nil {
				log.Printf("Failed to get attachments: %v", err)
			} else {
				fmt.Printf("\nFound %d attachment(s):\n", len(attachments))

				attachmentsDir := "attachments"
				if err := os.MkdirAll(attachmentsDir, 0o755); err != nil {
					log.Printf("Failed to create attachments directory: %v", err)
				}

				for i, att := range attachments {
					fmt.Println("───────────────────────────────────────────────────────────────")
					fmt.Printf("Attachment #%d:\n", i+1)
					fmt.Printf("  📄 Name: %s\n", att.Name)
					fmt.Printf("  📦 Type: %s\n", att.ContentType)
					fmt.Printf("  📊 Size: %d bytes\n", att.Size)
					fmt.Printf("  🔗 Is Inline: %v\n", att.IsInline)
					fmt.Printf("  🆔 ID: %s\n", att.ID)
					if att.ContentBytes != "" {
						fmt.Printf("  ✅ Content available (base64 encoded, %d chars)\n", len(att.ContentBytes))

						if !att.IsInline {
							filePath := filepath.Join(attachmentsDir, att.Name)
							if err := mailbox.SaveAttachment(att, filePath); err != nil {
								log.Printf("  ❌ Failed to save attachment: %v", err)
							} else {
								fmt.Printf("  💾 Saved to: %s\n", filePath)
							}
						}
					}
				}
			}
		} else {
			fmt.Println("\n📎 First message has no attachments.")
		}
	}

	// Fetch unread messages.
	fmt.Println("\n📨 Fetching unread messages...")
	unreadMessages, err := client.GetUnreadMessages(ctx, 10)
	if err != nil {
		log.Fatalf("Failed to get unread messages: %v", err)
	}

	fmt.Printf("\nFound %d unread messages:\n\n", len(unreadMessages))
	for _, msg := range unreadMessages {
		printMessage(msg)
	}
}
