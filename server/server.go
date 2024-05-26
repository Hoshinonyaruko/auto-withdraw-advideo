package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/structs"
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
func processWSMessage(msg []byte, config *config.Config) {
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

		//fmt.Printf("Processed a message event from group %d.\n", messageEvent.GroupID)

		// 提取视频URL
		re := regexp.MustCompile(`\[CQ:video,file=.+?,url=(.+?)\]`)
		matches := re.FindStringSubmatch(messageEvent.RawMessage)
		if len(matches) < 2 {
			//log.Println("No video URL found in the message.")
			return
		}
		videoURL := strings.Replace(matches[1], "\\u0026amp;", "&", -1)
		videoURL = strings.Replace(videoURL, "&amp;", "&", -1)
		fmt.Printf("提取到视频链接:%v\n", videoURL)
		encodedURL := url.QueryEscape(videoURL)

		// 拼接内部API请求
		port := config.Settings.Port
		selfID := fmt.Sprint(messageEvent.SelfID)
		messageID := fmt.Sprint(messageEvent.MessageID)
		apiURL := fmt.Sprintf("http://127.0.0.1:%s/videoDuration?videourl=%s&self_id=%s&message_id=%s", port, encodedURL, selfID, messageID)

		// 发起请求
		resp, err := http.Get(apiURL)
		if err != nil {
			log.Printf("Failed to invoke internal API: %v\n", err)
			return
		}
		defer resp.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		fmt.Println("Internal API response:", buf.String())
	} else {
		//log.Printf("Unknown message type or missing post type\n")
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
