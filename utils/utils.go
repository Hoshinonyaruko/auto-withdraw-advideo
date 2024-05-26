package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
)

type URLToken struct {
	BaseURL     string
	AccessToken string
}

var (
	baseURLMap   = make(map[string]URLToken)
	baseURLMapMu sync.Mutex
)

// 结构体用于解析 JSON 响应
type loginInfoResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		UserID   int64  `json:"user_id"`
		Nickname string `json:"nickname"`
	} `json:"data"`
}

// 构造 URL 并请求数据
func FetchAndStoreUserIDs() {
	accessTokens := config.GetHttpPathsAccessTokens() // 获取AccessToken配置

	httpPaths := config.GetHttpPaths()
	for _, baseURL := range httpPaths {
		url := baseURL + "/get_login_info"
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error fetching login info from %s: %v\n", url, err)
			continue
		}
		defer resp.Body.Close()

		var result loginInfoResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Printf("Error decoding response from %s: %v\n", url, err)
			continue
		}

		if result.Retcode == 0 && result.Status == "ok" {
			userIDStr := strconv.FormatInt(result.Data.UserID, 10)
			foundToken := "" // 默认token为空字符串

			// 检查是否存在对应的AccessToken
			for _, token := range accessTokens {
				if token.SelfID == userIDStr {
					foundToken = token.Token
					fmt.Printf("成功绑定机器人selfid[%v] onebot api baseURL[%v] with token[%v]\n", result.Data.UserID, baseURL, token.Token)
					break
				}
			}

			baseURLMapMu.Lock()
			baseURLMap[userIDStr] = URLToken{BaseURL: baseURL, AccessToken: foundToken}
			baseURLMapMu.Unlock()
			fmt.Printf("成功绑定机器人selfid[%v] onebot api baseURL[%v] without token\n", result.Data.UserID, baseURL)
		}
	}
}

// GetBaseURLByUserID 通过 user_id 获取 baseURL 和 accessToken
func GetBaseURLByUserID(userID string) (URLToken, bool) {
	baseURLMapMu.Lock()
	defer baseURLMapMu.Unlock()
	urlToken, exists := baseURLMap[userID]
	return urlToken, exists
}
