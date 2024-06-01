package webapi

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/logger"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/server"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/utils"
)

func GetVideoPlaylist(c *gin.Context) {
	videoURL := c.Query("videourl")
	selfID := c.Query("self_id")
	messageID := c.Query("message_id")
	userID := c.Query("user_id")
	GroupID := c.Query("group_id")

	if videoURL == "" || selfID == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "videourl, self_id, and message_id parameters are required"})
		return
	}

	decodedURL, err := url.QueryUnescape(videoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video URL"})
		return
	}

	duration, err := fetchVideoDuration(decodedURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("检测到视频,长度 %f\n", duration)

	videoSecondLimit := config.GetVideoSecondLimit()
	if duration < float64(videoSecondLimit) {
		fmt.Printf("Video duration %f is less than limit %d, deleting message_id %s for self_id %s\n", duration, videoSecondLimit, messageID, selfID)
		// 记录日志
		logger.LogEvent(fmt.Sprintf("Video duration %f is less than limit %d, deleting message_id %s for self_id %s", duration, videoSecondLimit, messageID, selfID))
		videopath := logger.DownloadVideo(decodedURL, selfID)

		if config.GetCheckVideoQRCode() {
			// 检查视频是否包含二维码
			if !utils.CheckVideoForQRCode(videopath) {
				fmt.Printf("video not contain QRcode pass.\n")
				logger.LogEvent(fmt.Sprintf("video not contain QRcode pass url:%s", decodedURL))
				c.JSON(http.StatusOK, gin.H{
					"duration": duration,
				})
				// 返回 不做处理
				return
			} else {
				fmt.Printf("video contain QRcode!!\n")
				logger.LogEvent(fmt.Sprintf("video contain QRcode!! url:%s", decodedURL))
				// 撤回 & 提示
			}
		}

		urlToken, exists := utils.GetBaseURLByUserID(selfID)
		if !exists {
			// Reverse WS request logic here
			server.SendDeleteMessageViaWebSocket(selfID, messageID)
			// 发提示
			server.SendGroupMessageViaWebSocket(selfID, GroupID, userID, config.GetWithdrawNotice())
		} else {
			server.SendDeleteRequest(urlToken, messageID)
			// 发提示
			server.SendGroupMsgHttp(urlToken, GroupID, userID, config.GetWithdrawNotice())
		}

		// 可以根据需要输出更多响应细节
		c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"duration": duration,
	})
}

func fetchVideoDuration(videoURL string) (float64, error) {
	// 创建自定义的HTTP客户端，忽略证书验证
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second, // 设置超时
	}

	// 使用自定义的客户端发起请求
	resp, err := client.Get(videoURL)
	if err != nil {
		return 0, fmt.Errorf("failed to get video: %v", err)
	}
	defer resp.Body.Close()

	buffer := make([]byte, 1024*1024*4) // 4MB buffer to find mvhd
	totalRead := 0
	for {
		n, err := resp.Body.Read(buffer[totalRead:])
		if err != nil {
			break
		}
		totalRead += n
		if totalRead >= len(buffer) {
			break
		}
	}

	mvhdIndex := findMvhd(buffer[:totalRead])
	if mvhdIndex == -1 {
		return 0, fmt.Errorf("mvhd box not found in the first %d bytes of the video", totalRead)
	}

	return parseDuration(buffer[mvhdIndex:])
}

func findMvhd(data []byte) int {
	const sizeOfLengthAndType = 8 // Length (4 bytes) + Type (4 bytes)
	index := 0

	for index+sizeOfLengthAndType <= len(data) {
		size := int(binary.BigEndian.Uint32(data[index : index+4]))
		boxType := string(data[index+4 : index+8])
		if boxType == "moov" {
			// Look for mvhd within moov
			endIndex := index + size
			index += sizeOfLengthAndType
			for index+sizeOfLengthAndType <= endIndex {
				subSize := int(binary.BigEndian.Uint32(data[index : index+4]))
				subBoxType := string(data[index+4 : index+8])
				if subBoxType == "mvhd" {
					return index
				}
				index += subSize
			}
		}
		index += size
	}
	return -1
}

func parseDuration(data []byte) (float64, error) {
	if len(data) < 24 {
		return 0, fmt.Errorf("insufficient data for duration calculation")
	}
	timeScale := binary.BigEndian.Uint32(data[20:24])
	duration := binary.BigEndian.Uint32(data[24:28])

	if timeScale == 0 {
		return 0, fmt.Errorf("invalid time scale value")
	}

	return float64(duration) / float64(timeScale), nil
}

// GetImageAndCheckQRCode handles the incoming request, downloads the image and checks for QR code.
func GetImageAndCheckQRCode(c *gin.Context) {
	imageURL := c.Query("imageurl")
	selfID := c.Query("self_id")
	messageID := c.Query("message_id")
	userID := c.Query("user_id")
	GroupID := c.Query("group_id")

	if imageURL == "" || selfID == "" || messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageurl, self_id, and message_id parameters are required"})
		return
	}

	// Create a custom HTTP client to ignore SSL certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	// Download the image
	resp, err := client.Get(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download image: %v", err)})
		return
	}
	defer resp.Body.Close()

	// Assuming the function to save the downloaded image and return the path
	imagePath, err := saveImage(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save image: %v", err)})
		return
	}

	// Check for QR code in the image
	if utils.ContainsQRCode(imagePath) {
		fmt.Println("Image contains a QR code.")

		urlToken, exists := utils.GetBaseURLByUserID(selfID)
		if !exists {
			// Reverse WS request logic here
			server.SendDeleteMessageViaWebSocket(selfID, messageID)
			// 发提示
			server.SendGroupMessageViaWebSocket(selfID, GroupID, userID, config.GetWithdrawNotice())
		} else {
			server.SendDeleteRequest(urlToken, messageID)
			// 发提示
			server.SendGroupMsgHttp(urlToken, GroupID, userID, config.GetWithdrawNotice())
		}

		c.JSON(http.StatusOK, gin.H{"message": "Image contains QR code, message deleted."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "No QR code detected in image.",
	})
}

// saveImage saves the image from the given reader to the local file system and returns the file path.
func saveImage(imageData io.Reader) (string, error) {
	// Generate a unique file name using UUID
	fileName := fmt.Sprintf("%s.jpg", uuid.New().String())
	filePath := filepath.Join("images", fileName) // Ensure the "images" directory exists

	// Create the "images" directory if it does not exist
	err := os.MkdirAll("images", os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy the image data to the file
	_, err = io.Copy(file, imageData)
	if err != nil {
		return "", fmt.Errorf("failed to save image: %v", err)
	}

	return filePath, nil
}
