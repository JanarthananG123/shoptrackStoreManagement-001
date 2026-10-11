package logger

import (
	"log"
	"os"
	"testing"
)

func TestLoggerInitAndClose(t *testing.T) {
	lFile, lErr := InitLogger()
	if lErr != nil {
		t.Fatalf("InitLogger failed: %v", lErr)
	}
	if lFile == nil {
		t.Fatalf("Expected non-nil log file")
	}

	log.Println("Test log entry")
	CloseLogger()

	lStat, lErr := os.Stat(lFile.Name())
	if lErr != nil {
		t.Fatalf("Log file does not exist: %v", lErr)
	}
	if lStat.Size() == 0 {
		t.Fatalf("Log file is empty")
	}
}
