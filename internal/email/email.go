package email

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	logger "email-to-printer/utils"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

var log *logger.Logger

func init() {
	log = logger.GetLogger()
}

type EmailConfig struct {
	Server   string
	Username string
	Password string
	Mailbox  string
}

type Attachment struct {
	Filename string
	FilePath string // Change back to FilePath instead of Data
}

// cleanupOldAttachments removes temporary files older than the specified duration
func cleanupOldAttachments(maxAge time.Duration) error {
	tempDir := filepath.Join(os.TempDir(), "email-attachments")
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read temp directory: %v", err)
	}

	now := time.Now()
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			fullPath := filepath.Join(tempDir, info.Name())
			if err := os.Remove(fullPath); err != nil {
				log.Error("Failed to remove old file %s: %v", fullPath, err)
			} else {
				log.Debug("Removed old temporary file: %s", fullPath)
			}
		}
	}
	return nil
}

// FetchUnseenEmails connects to IMAP, fetches unseen emails, and returns attachments
func FetchUnseenEmails(cfg *EmailConfig) ([]Attachment, error) {
	// Clean up old attachments (older than 24 hours)
	if err := cleanupOldAttachments(24 * time.Hour); err != nil {
		log.Error("Warning: Failed to cleanup old attachments: %v", err)
	}

	var attachments []Attachment
	var processedEmails int

	// Connect to IMAP server
	log.Info("Connecting to IMAP server: %s", cfg.Server)
	c, err := client.DialTLS(cfg.Server, nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %v", err)
	}
	defer c.Logout()

	// Login
	if err := c.Login(cfg.Username, cfg.Password); err != nil {
		return nil, fmt.Errorf("IMAP login failed: %v", err)
	}

	// Select mailbox
	_, err = c.Select(cfg.Mailbox, false)
	if err != nil {
		return nil, fmt.Errorf("mailbox selection failed: %v", err)
	}

	// Search for unseen messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	ids, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("message search failed: %v", err)
	}

	if len(ids) == 0 {
		log.Info("No new emails found")
		return attachments, nil
	}

	log.Info("Found %d new emails - starting batch processing", len(ids))

	// Process messages in batches of 10
	const batchSize = 10
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		log.Debug("Processing batch %d/%d (emails %d-%d of %d)",
			(i/batchSize)+1,
			(len(ids)+batchSize-1)/batchSize,
			i+1, end, len(ids))

		// Create sequence set for current batch
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids[i:end]...)

		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)

		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchRFC822}, messages)
		}()

		batchCount := 0
		// Process messages in current batch
		for msg := range messages {
			batchCount++
			processedEmails++
			log.Debug("Processing email %d in current batch", batchCount)

			section, _ := imap.ParseBodySectionName(imap.FetchRFC822)
			r := msg.GetBody(section)
			if r == nil {
				continue
			}

			// Create mail reader
			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Error("Failed to create mail reader: %v", err)
				continue
			}

			// Process message parts
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Error("Failed to read part: %v", err)
					continue
				}

				switch h := p.Header.(type) {
				case *mail.AttachmentHeader:
					filename, err := h.Filename()
					if err != nil || filename == "" {
						continue
					}

					// Create temp directory if not exists
					tempDir := filepath.Join(os.TempDir(), "email-attachments")
					if err := os.MkdirAll(tempDir, 0755); err != nil {
						log.Error("Failed to create temp dir: %v", err)
						continue
					}

					// Save attachment to temp file
					filePath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().Unix(), filename))
					f, err := os.Create(filePath)
					if err != nil {
						log.Error("Failed to create file: %v", err)
						continue
					}

					if _, err := io.Copy(f, p.Body); err != nil {
						f.Close()
						log.Error("Failed to save attachment: %v", err)
						continue
					}
					f.Close()

					attachments = append(attachments, Attachment{
						Filename: filename,
						FilePath: filePath,
					})

					log.Debug("Saved attachment: %s", filePath)
				}
			}
		}

		// Wait for fetch to complete and check for errors
		if err := <-done; err != nil {
			log.Error("Failed to fetch messages in batch: %v", err)
			continue
		}
		log.Debug("Successfully processed %d emails in current batch", batchCount)

		// Mark messages as read
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqset, item, flags, nil); err != nil {
			log.Error("Failed to mark messages as read: %v", err)
		} else {
			log.Debug("Marked %d messages as read", end-i)
		}
	}

	log.Info("Email processing complete - Processed %d emails, found %d attachments",
		processedEmails,
		len(attachments))

	return attachments, nil
}

// ValidateAttachment checks if file type is allowed
func ValidateAttachment(filename string, allowedTypes []string) bool {
	if len(allowedTypes) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, t := range allowedTypes {
		if ext == strings.ToLower(t) {
			return true
		}
	}
	return false
}
