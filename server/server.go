package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/logger"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/structs"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/superini"
)

type WebSocketServerClient struct {
	SelfID string
	Conn   *websocket.Conn
}

var lock sync.Mutex

// 维护所有活跃连接的切片
var clients = []*WebSocketServerClient{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 使用闭包结构 因为gin需要c *gin.Context固定签名
func WsHandlerWithDependencies(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		wsHandler(c, config)
	}
}

// 处理正向ws客户端的连接
func wsHandler(c *gin.Context, config *config.Config) {
	// 先从请求头中尝试获取token
	tokenFromHeader := c.Request.Header.Get("Authorization")
	selfID := c.Request.Header.Get("X-Self-ID")
	fmt.Printf("接入机器人X-Self-ID[%v]", selfID)

	token := ""
	if tokenFromHeader != "" {
		if strings.HasPrefix(tokenFromHeader, "Token ") {
			// 从 "Token " 后面提取真正的token值
			token = strings.TrimPrefix(tokenFromHeader, "Token ")
		} else if strings.HasPrefix(tokenFromHeader, "Bearer ") {
			// 从 "Bearer " 后面提取真正的token值
			token = strings.TrimPrefix(tokenFromHeader, "Bearer ")
		} else {
			// 直接使用token值
			token = tokenFromHeader
		}
	} else {
		// 如果请求头中没有token，则从URL参数中获取
		token = c.Query("access_token")
	}

	// 获取配置中的有效 token
	validToken := config.Settings.Wstoken

	// 如果配置的 token 不为空，但提供的 token 为空或不匹配
	if validToken != "" && (token == "" || token != validToken) {
		if token == "" {
			fmt.Printf("Connection failed due to missing token. Headers: %v", c.Request.Header)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		} else {
			fmt.Printf("Connection failed due to incorrect token. Headers: %v, Provided token: %s", c.Request.Header, token)
			c.JSON(http.StatusForbidden, gin.H{"error": "Incorrect token"})
		}
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v", err)
		return
	}

	clientIP := c.ClientIP()
	fmt.Printf("WebSocket client connected. IP: %s", clientIP)

	// 创建WebSocketServerClient实例
	client := &WebSocketServerClient{
		Conn: conn,
	}

	botID := selfID

	lock.Lock()
	clients = append(clients, &WebSocketServerClient{
		SelfID: selfID,
		Conn:   conn,
	})
	lock.Unlock()

	// 发送连接成功的消息
	message := map[string]interface{}{
		"meta_event_type": "lifecycle",
		"post_type":       "meta_event",
		"self_id":         botID,
		"sub_type":        "connect",
		"time":            int(time.Now().Unix()),
	}
	err = client.SendMessage(message)
	if err != nil {
		fmt.Printf("Error sending connection success message: %v\n", err)
	}

	//退出时候的清理
	defer conn.Close()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %v", err)
			return
		}

		if messageType == websocket.TextMessage {
			processWSMessage(p, config)
		}
	}
}

// 处理收到的信息
func processWSMessage(msg []byte, conf *config.Config) {
	var genericMap map[string]interface{}
	if err := json.Unmarshal(msg, &genericMap); err != nil {
		log.Printf("Error unmarshalling message to map: %v, Original message: %s\n", err, string(msg))
		return
	}

	if postType, ok := genericMap["post_type"].(string); ok && postType == "message" {
		var messageEvent structs.MessageEvent
		if err := json.Unmarshal(msg, &messageEvent); err != nil {
			log.Printf("Error unmarshalling message event: %v\n", err)
			return
		}

		rawMessage := messageEvent.RawMessage
		groupID := fmt.Sprint(messageEvent.GroupID)
		selfID := fmt.Sprint(messageEvent.SelfID)
		userID := fmt.Sprint(messageEvent.UserID)
		messageID := fmt.Sprint(messageEvent.MessageID)

		// 获取需要撤回的关键词列表
		withdrawWords := config.GetWithdrawWords()

		// 检查rawMessage是否包含任何撤回关键词
		for _, word := range withdrawWords {
			if strings.Contains(rawMessage, word) {
				// 如果找到匹配的词，则发送删除消息请求
				SendDeleteMessageViaWebSocket(selfID, messageID)
				// 撤回 & 提示
				logger.LogEvent(fmt.Sprintf("bot [%s] withdraw from group_id:%s user_id:%s messgae[%s]", selfID, groupID, userID, rawMessage))
				// 发送提示消息
				withdrawNotice := config.GetWithdrawNotice()
				SendGroupMessageViaWebSocket(selfID, groupID, userID, withdrawNotice)

				// 如果设置了踢出群成员
				if config.GetSetGroupKick() {
					KickGroupMemberViaWebSocket(selfID, groupID, userID)
				}

				// 处理完毕后退出循环
				break
			}
		}

		handleConfigToggle := func(currentStatus string, enableMessage, disableMessage string, section string) {
			newStatus := "true"
			if currentStatus == "true" {
				newStatus = "false"
			}
			superini.WriteConfig(groupID, section, newStatus)
			message := disableMessage
			if newStatus == "true" {
				message = enableMessage
			}
			SendGroupMessageViaWebSocket(selfID, groupID, userID, message)
		}

		switch rawMessage {
		case config.GetOnEnableVideoCheck():
			currentStatus := superini.ReadConfig(groupID, "handleVideoMessage")
			handleConfigToggle(currentStatus, "视频二维码检测已开启", "视频二维码检测已关闭", "handleVideoMessage")

		case config.GetOnDisableVideoCheck():
			currentStatus := superini.ReadConfig(groupID, "handleVideoMessage")
			handleConfigToggle(currentStatus, "视频二维码检测已开启", "视频二维码检测已关闭", "handleVideoMessage")

		case config.GetOnEnablePicCheck():
			currentStatus := superini.ReadConfig(groupID, "handleImageMessage")
			handleConfigToggle(currentStatus, "图片二维码检测已开启", "图片二维码检测已关闭", "handleImageMessage")

		case config.GetOnDisablePicCheck():
			currentStatus := superini.ReadConfig(groupID, "handleImageMessage")
			handleConfigToggle(currentStatus, "图片二维码检测已开启", "图片二维码检测已关闭", "handleImageMessage")

		default:
			videoCheckEnabled := superini.ReadConfig(groupID, "handleVideoMessage")
			imageCheckEnabled := superini.ReadConfig(groupID, "handleImageMessage")

			if strings.Contains(rawMessage, "[CQ:video") && videoCheckEnabled == "true" {
				handleVideoMessage(conf, rawMessage, messageEvent)
			} else if strings.Contains(rawMessage, "[CQ:image") && imageCheckEnabled == "true" {
				handleImageMessage(conf, rawMessage, messageEvent)
			}
		}
	}
}

// 发信息给client
func (c *WebSocketServerClient) SendMessage(message map[string]interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return err
	}
	return c.Conn.WriteMessage(websocket.TextMessage, msgBytes)
}

func (client *WebSocketServerClient) Close() error {
	return client.Conn.Close()
}

// 发信息给client
func SendMessageBySelfID(selfID string, message map[string]interface{}) error {
	lock.Lock()
	defer lock.Unlock()

	for _, client := range clients {
		if client.SelfID == selfID {
			msgBytes, err := json.Marshal(message)
			if err != nil {
				return fmt.Errorf("error marshalling message: %v", err)
			}
			return client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
		}
	}

	return fmt.Errorf("no connection found for selfID: %s", selfID)
}
