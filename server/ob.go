package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// sendDeleteMessageViaWebSocket sends a delete message via WebSocket for the given selfID and messageID.
func SendDeleteMessageViaWebSocket(selfID, messageID string) error {
	message := map[string]interface{}{
		"action": "delete_msg",
		"params": map[string]string{
			"message_id": messageID,
		},
	}

	if err := SendMessageBySelfID(selfID, message); err != nil {
		log.Printf("Failed to send delete message via websocket: %v\n", err)
		return err
	}

	return nil
}

// sendDeleteRequest sends a delete request via HTTP to the given URL token.
func SendDeleteRequest(urlToken struct {
	BaseURL     string
	AccessToken string
}, messageID string) error {
	deleteURL := fmt.Sprintf("%s/delete_msg?message_id=%s", urlToken.BaseURL, messageID)
	if urlToken.AccessToken != "" {
		deleteURL += fmt.Sprintf("&access_token=%s", urlToken.AccessToken)
	}

	resp, err := http.Get(deleteURL)
	if err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete message, status code: %d", resp.StatusCode)
	}

	return nil
}

// sendGroupMsgHttp sends a POST request to the fixed endpoint with the given group_id, user_id and message.
func SendGroupMsgHttp(urlToken struct {
	BaseURL     string
	AccessToken string
}, groupID, userID, message string) error {
	// 定义固定的endpoint
	const endpoint = "send_group_msg"

	// 构建完整的URL
	postURL := fmt.Sprintf("%s/%s", urlToken.BaseURL, endpoint)
	u, err := url.Parse(postURL)
	if err != nil {
		return fmt.Errorf("URL parsing failed: %v", err)
	}

	// 添加access_token参数
	query := u.Query()
	if urlToken.AccessToken != "" {
		query.Set("access_token", urlToken.AccessToken)
	}
	u.RawQuery = query.Encode()

	// 构造请求体
	payload := map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
		"message":  "[CQ:at,qq=" + userID + "]" + message, //at触发者
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return nil
}

// SendGroupMessageViaWebSocket sends a group message via WebSocket for the given selfID, groupID, userID, and message.
func SendGroupMessageViaWebSocket(selfID, groupID, userID, message string) error {
	wsMessage := map[string]interface{}{
		"action": "send_group_msg",
		"params": map[string]interface{}{
			"group_id": groupID,
			"user_id":  userID,
			"message":  message,
		},
	}

	if err := SendMessageBySelfID(selfID, wsMessage); err != nil {
		log.Printf("Failed to send group message via websocket: %v\n", err)
		return err
	}

	return nil
}
