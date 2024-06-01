package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/server"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/superini"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/template"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/utils"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/webapi"
)

func main() {
	// 如果用户指定了-yml参数
	configFilePath := "config.yml" // 默认配置文件路径

	// 检查配置文件是否存在
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		handleMissingConfigFile(configFilePath)
	}

	// 加载配置
	superini.GetInstance()
	conf, err := config.LoadConfig(configFilePath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// 判断是否设置多个http地址,获取对应关系
	if len(config.GetHttpPaths()) > 0 {
		utils.FetchAndStoreUserIDs()
	}
	router := gin.Default()
	router.GET("/videoDuration", webapi.GetVideoPlaylist)
	router.GET("/picheck", webapi.GetImageAndCheckQRCode)

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

	// 调试用
	//bool := utils.ContainsQRCode("C:\\Users\\Cosmo\\Pictures\\haibao.png")
	//bool := utils.ContainsQRCode("C:\\Users\\Cosmo\\Pictures\\frame_0002.jpg")
	//bool := utils.ContainsQRCode("C:\\Users\\Cosmo\\Pictures\\智能体背景1.jpg")
	//bool := utils.ContainsQRCode("C:\\Users\\Cosmo\\Pictures\\支付宝.JPG")

	//fmt.Printf("%v", bool)

	// 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	// 可以执行退出程序
	// 正常退出程序
	os.Exit(0)

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
