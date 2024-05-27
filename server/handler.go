package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/structs"
)

func handleVideoMessage(config *config.Config, rawMessage string, messageEvent structs.MessageEvent) {
	re := regexp.MustCompile(`\[CQ:video,file=.+?,url=(.+?)\]`)
	matches := re.FindStringSubmatch(rawMessage)
	if len(matches) < 2 {
		// log.Println("No video URL found in the message.")
		return
	}

	videoURL := strings.Replace(matches[1], "\\u0026amp;", "&", -1)
	videoURL = strings.Replace(videoURL, "&amp;", "&", -1)
	fmt.Printf("提取到视频链接:%v\n", videoURL)
	encodedURL := url.QueryEscape(videoURL)

	port := config.Settings.Port
	selfID := fmt.Sprint(messageEvent.SelfID)
	messageID := fmt.Sprint(messageEvent.MessageID)
	userID := fmt.Sprint(messageEvent.UserID)
	groupID := fmt.Sprint(messageEvent.GroupID)
	apiURL := fmt.Sprintf("http://127.0.0.1:%s/videoDuration?videourl=%s&self_id=%s&message_id=%s&user_id=%s&group_id=%s", port, encodedURL, selfID, messageID, userID, groupID)

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Failed to invoke internal API: %v\n", err)
		return
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	fmt.Println("Internal API response:", buf.String())
}

func handleImageMessage(config *config.Config, rawMessage string, messageEvent structs.MessageEvent) {
	re := regexp.MustCompile(`\[CQ:image,file=.+?,url=(.+?)\]`)
	matches := re.FindAllStringSubmatch(rawMessage, -1)
	if len(matches) == 0 {
		// log.Println("No image URL found in the message.")
		return
	}

	port := config.Settings.Port
	selfID := fmt.Sprint(messageEvent.SelfID)
	messageID := fmt.Sprint(messageEvent.MessageID)
	userID := fmt.Sprint(messageEvent.UserID)
	groupID := fmt.Sprint(messageEvent.GroupID)

	for _, match := range matches {
		imageURL := strings.Replace(match[1], "\\u0026amp;", "&", -1)
		imageURL = strings.Replace(imageURL, "&amp;", "&", -1)
		fmt.Printf("提取到图片链接:%v\n", imageURL)
		encodedURL := url.QueryEscape(imageURL)

		apiURL := fmt.Sprintf("http://127.0.0.1:%s/picheck?imageurl=%s&self_id=%s&message_id=%s&user_id=%s&group_id=%s", port, encodedURL, selfID, messageID, userID, groupID)

		resp, err := http.Get(apiURL)
		if err != nil {
			log.Printf("Failed to invoke internal API: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		fmt.Println("Internal API response:", buf.String())
	}
}
