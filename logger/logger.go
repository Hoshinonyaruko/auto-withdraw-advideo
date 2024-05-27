package logger

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var logFolder = "video"

func init() {
	// 确保日志文件夹存在
	err := os.MkdirAll(logFolder, 0755)
	if err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}
}

// LogEvent logs the specified message to a file named with today's date
func LogEvent(message string) {
	now := time.Now()
	filename := filepath.Join(logFolder, now.Format("2006-01-02")+".log")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	logMessage := fmt.Sprintf("%s: %s\n", now.Format("15:04:05"), message)
	if _, err = file.WriteString(logMessage); err != nil {
		log.Fatalf("Failed to write to log file: %v", err)
	}
}

// DownloadVideo downloads the video from the specified URL and ignores SSL certificate errors
func DownloadVideo(url, selfID string) string {
	// Create a custom HTTP client that ignores certificate errors
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	// Use the custom client to initiate a request
	resp, err := client.Get(url)
	if err != nil {
		LogEvent(fmt.Sprintf("Failed to download video for selfID %s: %v", selfID, err))
		return ""
	}
	defer resp.Body.Close()

	hash := md5.Sum([]byte(url))
	filePath := filepath.Join(logFolder, fmt.Sprintf("%x.mp4", hash))
	file, err := os.Create(filePath)
	if err != nil {
		LogEvent(fmt.Sprintf("Failed to create file for video URL %s: %v", url, err))
		return ""
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		LogEvent(fmt.Sprintf("Failed to save video for URL %s: %v", url, err))
		return ""
	}

	LogEvent(fmt.Sprintf("Video downloaded successfully for URL %s, saved to %s", url, filePath))
	return filePath
}
