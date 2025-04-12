package printer

import (
	logger "email-to-printer/utils"
	"os"
	"os/exec"
	"path/filepath"
)

var log *logger.Logger

func init() {
	log = logger.GetLogger()
}

func PrintFile(filepath string, printerName string) error {
	cmd := exec.Command("lp", "-d", printerName, filepath)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Error("Print failed: %v, output: %s", err, string(output))
		return err
	}
	log.Info("Successfully printed file: %s", filepath)
	return nil
}

// PrintData saves data to temp file and prints it using PrintFile
func PrintData(data []byte, filename string, printerName string) error {
	tempDir := filepath.Join(os.TempDir(), "email-printer")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Error("Failed to create temp directory: %v", err)
		return err
	}

	tempFile := filepath.Join(tempDir, filename)
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		log.Error("Failed to write temp file: %v", err)
		return err
	}
	defer os.Remove(tempFile)

	return PrintFile(tempFile, printerName)
}
