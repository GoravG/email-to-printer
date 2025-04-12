package main

import (
	"os"

	"email-to-printer/config"
	"email-to-printer/internal/email"
	"email-to-printer/internal/printer"

	logger "email-to-printer/utils"
)

func main() {
	log := logger.GetLogger()
	log.Info("Starting email-to-printer application")

	cfg := config.Load()
	log.Debug("Configuration loaded: %+v", cfg)

	emailCfg := &email.EmailConfig{
		Server:   cfg.ImapServer,
		Username: cfg.Email,
		Password: cfg.Password,
		Mailbox:  "INBOX",
	}

	log.Info("Fetching emails...")
	attachments, err := email.FetchUnseenEmails(emailCfg)
	if err != nil {
		log.Error("Failed to fetch emails: %v", err)
		os.Exit(1)
	}
	log.Info("Successfully fetched %d attachments", len(attachments))

	for _, att := range attachments {
		log.Debug("Processing attachment: %s", att.Filename)
		if err := printer.PrintFile(att.FilePath, cfg.PrinterName); err != nil {
			log.Error("Failed to print %s: %v", att.Filename, err)
		} else {
			log.Info("Successfully printed file: %s", att.Filename)
		}
	}

	log.Info("Email-to-printer application completed")
}
