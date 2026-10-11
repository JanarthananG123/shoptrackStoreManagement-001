package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var gLogFile *os.File

// InitLogger initializes application logging with dual console and microsecond-named file output.
func InitLogger() (*os.File, error) {
	log.Println("[Logger][InitLogger][IN]")
	defer log.Println("[Logger][InitLogger][OUT]")

	lLogDir := "logs"
	// Ensure logs are consistently placed in the Backend root even if executed from cmd/server
	if _, lErr := os.Stat("go.mod"); os.IsNotExist(lErr) {
		if _, lErr2 := os.Stat("../../go.mod"); lErr2 == nil {
			lLogDir = "../../logs"
		} else if _, lErr3 := os.Stat("../go.mod"); lErr3 == nil {
			lLogDir = "../logs"
		}
	}

	if lErr := os.MkdirAll(lLogDir, 0755); lErr != nil {
		lWrapErr := fmt.Errorf("failed to create log directory: %w", lErr)
		log.Printf("[Logger][InitLogger][ERROR] %v", lWrapErr)
		return nil, lWrapErr
	}

	lTimestamp := time.Now().Format("2006-01-02_15-04-05.000000")
	lFileName := filepath.Join(lLogDir, fmt.Sprintf("app_%s.log", lTimestamp))

	lFile, lErr := os.OpenFile(lFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if lErr != nil {
		lWrapErr := fmt.Errorf("failed to open log file %s: %w", lFileName, lErr)
		log.Printf("[Logger][InitLogger][ERROR] %v", lWrapErr)
		return nil, lWrapErr
	}

	gLogFile = lFile
	lMultiWriter := io.MultiWriter(os.Stdout, lFile)
	log.SetOutput(lMultiWriter)
	log.SetFlags(log.Ldate | log.Ltime)

	log.Printf("[Logger][InitLogger] Logging initialized successfully. Output target: %s", lFileName)
	return lFile, nil
}

// CloseLogger flushes and closes the active file handler on server exit.
func CloseLogger() {
	log.Println("[Logger][CloseLogger][IN]")
	defer log.Println("[Logger][CloseLogger][OUT]")

	if gLogFile != nil {
		if lErr := gLogFile.Sync(); lErr != nil {
			log.Printf("[Logger][CloseLogger][ERROR] Failed to sync log file: %v", lErr)
		}
		if lErr := gLogFile.Close(); lErr != nil {
			log.Printf("[Logger][CloseLogger][ERROR] Failed to close log file: %v", lErr)
		}
		gLogFile = nil
	}
}
