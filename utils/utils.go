package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/config"
	"github.com/hoshinonyaruko/auto-withdraw-advideo/logger"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
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

func CheckVideoForQRCode(videoPath string) bool {
	framesDir := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	if err := os.MkdirAll(framesDir, os.ModePerm); err != nil {
		logger.LogEvent(fmt.Sprintf("Failed to create directory for frames: %v", err))
		return false
	}

	if err := extractFrames(videoPath, framesDir); err != nil {
		logger.LogEvent(fmt.Sprintf("Frame extraction failed: %v", err))
		return false
	}

	// Scan frames for QR codes
	qrCount := 0
	frameFiles, err := filepath.Glob(filepath.Join(framesDir, "*.jpg"))
	if err != nil {
		logger.LogEvent(fmt.Sprintf("Failed to list frame files: %v", err))
		return false
	}
	qrlimit := config.GetQRLimit()

	for _, frame := range frameFiles {
		if ContainsQRCode(frame) {
			fmt.Printf("检测到视频帧包含二维码!\n")
			qrCount++
			if qrCount >= qrlimit { // Consider making this a variable or configurable
				return true
			}
		} else {
			fmt.Printf("未检测到视频帧包含二维码\n")
		}
	}
	return false
}

func ContainsQRCode(framePath string) bool {
	file, err := os.Open(framePath)
	if err != nil {
		log.Printf("Failed to open frame file: %v", err)
		return false
	}
	defer file.Close()

	img, err := imaging.Decode(file)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return false
	}

	// 图像预处理：转换为灰度图像
	grayImg := imaging.Grayscale(img)

	// 定义裁剪区域的百分比
	cropRatios := []struct {
		top    float32
		bottom float32
	}{
		{0.0, 0.0}, // 完整图像
		{0.2, 0.0}, // 上方裁剪20%，下方不裁剪
		{0.3, 0.0}, // 上方裁剪30%，下方不裁剪
		{0.4, 0.0}, // 上方裁剪40%，下方不裁剪
		{0.0, 0.2}, // 上方不裁剪，下方裁剪20%
		{0.0, 0.3}, // 上方不裁剪，下方裁剪30%
		{0.0, 0.4}, // 上方不裁剪，下方裁剪40%
		{0.1, 0.1}, // 上下都裁剪
		{0.2, 0.2}, // 上下都裁剪
		{0.3, 0.3}, // 上下都裁剪
		{0.4, 0.4}, // 上下都裁剪
		{0.5, 0.5}, // 上下都裁剪
	}

	// 尝试解码不同裁剪的图像
	for _, ratio := range cropRatios {
		croppedImg := cropImage(grayImg, ratio.top, ratio.bottom)
		if tryDecodeQRCode(croppedImg) {
			return true
		}
		if detectQRCodePresence(croppedImg) {
			return true
		}
	}

	return false
}

// 裁剪图像
func cropImage(img image.Image, topCrop float32, bottomCrop float32) image.Image {
	bounds := img.Bounds()
	height := bounds.Dy()
	cropTop := int(float32(height) * topCrop)
	cropBottom := int(float32(height) * bottomCrop)
	return imaging.Crop(img, image.Rect(0, cropTop, bounds.Dx(), height-cropBottom))
}

// 尝试解码QR码
func tryDecodeQRCode(img image.Image) bool {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		log.Printf("Failed to create binary bitmap: %v", err)
		return false
	}

	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER:       true,
		gozxing.DecodeHintType_PURE_BARCODE:     true, // 假设图像是纯粹的二维码
		gozxing.DecodeHintType_POSSIBLE_FORMATS: []gozxing.BarcodeFormat{gozxing.BarcodeFormat_QR_CODE},
	}
	qrReader := qrcode.NewQRCodeReader()
	_, err = qrReader.Decode(bmp, hints)
	if err != nil {
		// 检查错误消息中是否包含边界信息
		if matchBoundaryInfo(err.Error()) {
			log.Printf("QR code boundaries detected: %v", err)
			return true
		}
		log.Printf("Failed to decode QR code: %v", err)
		return false
	}

	return true
}

// 检测错误信息中是否含有可能表示二维码存在的信息
func matchQRCodeErrorInfo(errorMessage string) bool {
	// 匹配可能表示二维码部分成功识别的错误消息
	var re = regexp.MustCompile(`FormatException|ChecksumException|ReedSolomonException`)
	return re.MatchString(errorMessage)
}

// detectQRCodePresence tries to detect a QR Code in the given image.
func detectQRCodePresence(img image.Image) bool {
	source := gozxing.NewLuminanceSourceFromImage(img)
	var zz gozxing.GlobalHistogramBinarizer
	binarizer := zz.CreateBinarizer(source)
	bmp, err := gozxing.NewBinaryBitmap(binarizer)
	if err != nil {
		log.Printf("Failed to create binary bitmap: %v", err)
		return false
	}

	qrReader := qrcode.NewQRCodeReader()
	_, err = qrReader.Decode(bmp, nil)
	if err != nil {
		log.Printf("Failed to detect QR code: %v", err)
		// 检查错误信息是否包含可能表明二维码存在的信息
		if matchQRCodeErrorInfo(err.Error()) {
			log.Printf("Potential QR Code features detected despite the error: %v", err)
			return true
		}
		return false
	}

	log.Printf("QR Code successfully detected.")
	return true
}

// 检测错误信息中是否含有边界位置信息
func matchBoundaryInfo(errorMessage string) bool {
	// 正则表达式匹配类似 "(left,right)=(287,241), (top,bottom)=(44,1022)" 的信息
	re := regexp.MustCompile(`\(left,right\)=\((\d+),(\d+)\), \(top,bottom\)=\((\d+),(\d+)\)`)
	return re.MatchString(errorMessage)
}

func extractFrames(videoPath, framesDir string) error {
	// 确保 videoPath 是绝对路径
	absVideoPath, err := filepath.Abs(videoPath)
	if err != nil {
		fmt.Printf("Failed to get absolute path for video: %v\n", err)
		return err
	}

	// 确保 framesDir 也是绝对路径
	absFramesDir, err := filepath.Abs(framesDir)
	if err != nil {
		fmt.Printf("Failed to get absolute path for frames directory: %v\n", err)
		return err
	}

	// 检查并创建目录
	if err := os.MkdirAll(absFramesDir, 0755); err != nil {
		fmt.Printf("Failed to create frames directory: %v\n", err)
		return err
	}

	// 构建 ffmpeg 命令
	cmd := exec.Command("ffmpeg", "-i", absVideoPath, "-vf", "fps=1", filepath.Join(absFramesDir, "frame_%04d.jpg"))
	cmd.Dir = absFramesDir

	// 捕获标准输出和标准错误
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 运行命令
	err = cmd.Run()
	if err != nil {
		// 记录标准输出和标准错误
		fmt.Printf("ffmpeg stdout: %s\n", stdout.String())
		fmt.Printf("ffmpeg stderr: %s\n", stderr.String())
		fmt.Printf("ffmpeg command failed: %v\n", err)
		return fmt.Errorf("ffmpeg command failed with error: %v, stdout: %s, stderr: %s", err, stdout.String(), stderr.String())
	}
	return nil
}
