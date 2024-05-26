package main

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/logger"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/server"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/template"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/utils"
)

func main() {
	// 如果用户指定了-yml参数
	configFilePath := "config.yml" // 默认配置文件路径

	// 检查配置文件是否存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		handleMissingConfigFile(configFilePath)
	}

	// 加载配置
	conf, err := config.LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 判断是否设置多个http地址,获取对应关系
	if len(config.GetHttpPaths()) > 0 {
		utils.FetchAndStoreUserIDs()
	}
	router := gin.Default()
	router.GET("/videoDuration", getVideoPlaylist)
	//正向ws

	wspath := conf.Settings.WsPath
	if wspath == "nil" {
		router.GET("", server.WsHandlerWithDependencies(conf))
		fmt.Println("正向ws启动成功,监听0.0.0.0:" + conf.Settings.Port + "请注意设置ws_server_token(可空),并对外放通端口...")
	} else {
		router.GET("/"+wspath, server.WsHandlerWithDependencies(conf))
		fmt.Println("正向ws启动成功,监听0.0.0.0:" + conf.Settings.Port + "/" + wspath + "请注意设置ws_server_token(可空),并对外放通端口...")
	}

	// 创建一个http.Server实例(主服务器)
	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + conf.Settings.Port,
		Handler: router,
	}

	fmt.Printf("webui-api运行在 HTTP 端口 %v\n", conf.Settings.Port)
	// 在一个新的goroutine中启动主服务器
	go func() {
		// 使用HTTP
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	// 可以执行退出程序
	// 正常退出程序
	os.Exit(0)

}

func getVideoPlaylist(c *gin.Context) {
	videoURL := c.Query("videourl")
	selfID := c.Query("self_id")
	messageID := c.Query("message_id")

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
		logger.DownloadVideo(decodedURL, selfID)

		urlToken, exists := utils.GetBaseURLByUserID(selfID)
		if !exists {
			// 走反向ws请求
			message := map[string]interface{}{
				"action": "delete_msg",
				"params": map[string]string{
					"message_id": messageID,
				},
			}

			if err := server.SendMessageBySelfID(selfID, message); err != nil {
				log.Printf("Failed to send delete message via websocket: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send delete command via websocket"})
				return
			}

		} else {
			deleteURL := fmt.Sprintf("%s/delete_msg?message_id=%s", urlToken.BaseURL, messageID)
			if urlToken.AccessToken != "" {
				deleteURL += fmt.Sprintf("&access_token=%s", urlToken.AccessToken)
			}

			resp, err := http.Get(deleteURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete message: %v", err)})
				return
			}
			defer resp.Body.Close()
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

func handleMissingConfigFile(configFilePath string) {

	// 用户没有指定-yml参数，按照默认行为处理
	err := os.WriteFile(configFilePath, []byte(template.ConfigTemplate), 0644)
	if err != nil {
		fmt.Println("Error writing config.yml:", err)
		return
	}
	fmt.Println("请配置config.yml然后再次运行.")
	fmt.Print("按下 Enter 继续...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	os.Exit(0)

}
