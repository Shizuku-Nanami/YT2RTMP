package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"syscall"
	"time"
	"os"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
)

// 定义请求结构
type StreamRequest struct {
	URL       string `json:"url"`
	RTMPURL   string `json:"rtmp_url"`
	StreamKey string `json:"stream_key"`
}

// 定义流状态结构
type StreamStatus struct {
	IsRunning bool   `json:"is_running"`
	ProcessID int    `json:"process_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// 定义流进程结构
type StreamProcess struct {
	Process   *exec.Cmd
	RTMPURL   string
	StartTime time.Time
}

// 全局变量
var (
	streams = make(map[string]*StreamProcess)
	mu      sync.RWMutex
	server  *socketio.Server
)

func init() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	server = socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Printf("Client connected: %s", s.ID())
		return nil
	})

	server.OnEvent("/", "disconnect", func(s socketio.Conn, reason string) {
		log.Printf("Client disconnected: %s, reason: %s", s.ID(), reason)
	})
}

// 广播流状态更新
func broadcastStreamsUpdate() {
	mu.RLock()
	streamsList := make([]map[string]interface{}, 0)
	for url, process := range streams {
		status := "stopped"
		if process.Process.Process != nil && process.Process.ProcessState == nil {
			status = "running"
		}
		streamsList = append(streamsList, map[string]interface{}{
			"url":    url,
			"status": status,
			"pid":    process.Process.Process.Pid,
		})
	}
	mu.RUnlock()

	server.BroadcastToNamespace("/", "streams_update", map[string]interface{}{
		"streams": streamsList,
	})
}

// 启动流
func startStream(c *gin.Context) {
	var request StreamRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rtmpURL := fmt.Sprintf("%s/%s", request.RTMPURL, request.StreamKey)
	ffmpegCmd := fmt.Sprintf(`ffmpeg -re -i "$(yt-dlp -g --cookies "%s" %s)" -c:v copy -c:a copy -f flv "%s"`, os.Getenv("COOKIES_PATH"), request.URL, rtmpURL)

	cmd := exec.Command("bash", "-c", ffmpegCmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	streams[request.URL] = &StreamProcess{
		Process:   cmd,
		RTMPURL:   rtmpURL,
		StartTime: time.Now(),
	}
	mu.Unlock()

	// 启动监控协程
	go monitorProcess(request.URL, cmd)

	// 广播更新
	broadcastStreamsUpdate()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "推流已启动",
		"pid":     cmd.Process.Pid,
	})
}

// 停止流
func stopStream(c *gin.Context) {
	url := c.Query("url")
	mu.Lock()
	process, exists := streams[url]
	mu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到对应的推流进程"})
		return
	}

	// 终止进程组
	if err := syscall.Kill(-process.Process.Process.Pid, syscall.SIGTERM); err != nil {
		log.Printf("Error sending SIGTERM: %v", err)
	}

	// 等待一小段时间
	time.Sleep(time.Second)

	// 检查进程是否还在运行
	if process.Process.ProcessState == nil {
		// 强制终止
		if err := syscall.Kill(-process.Process.Process.Pid, syscall.SIGKILL); err != nil {
			log.Printf("Error sending SIGKILL: %v", err)
		}
	}

	mu.Lock()
	delete(streams, url)
	mu.Unlock()

	// 广播更新
	broadcastStreamsUpdate()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "推流已停止",
	})
}

// 获取流状态
func getStreamStatus(c *gin.Context) {
	url := c.Param("url")
	mu.RLock()
	process, exists := streams[url]
	mu.RUnlock()

	if !exists {
		c.JSON(http.StatusOK, StreamStatus{
			IsRunning: false,
			Error:     "未找到对应的推流进程",
		})
		return
	}

	status := StreamStatus{
		IsRunning: process.Process.ProcessState == nil,
		ProcessID: process.Process.Process.Pid,
	}

	c.JSON(http.StatusOK, status)
}

// 列出所有流
func listStreams(c *gin.Context) {
	mu.RLock()
	streamsList := make([]map[string]interface{}, 0)
	for url, process := range streams {
		status := "stopped"
		if process.Process.ProcessState == nil {
			status = "running"
		}
		streamsList = append(streamsList, map[string]interface{}{
			"url":    url,
			"status": status,
			"pid":    process.Process.Process.Pid,
		})
	}
	mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{"streams": streamsList})
}

// 监控进程
func monitorProcess(url string, cmd *exec.Cmd) {
	cmd.Wait()
	mu.Lock()
	delete(streams, url)
	mu.Unlock()
	broadcastStreamsUpdate()
}

func main() {
	r := gin.Default()

	// 配置 CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("DOMAIN_URL"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Socket.IO 路由
	r.GET("/socket.io/*any", gin.WrapH(server))
	r.POST("/socket.io/*any", gin.WrapH(server))

	// API 路由
	r.POST("/api/stream/start", startStream)
	r.POST("/api/stream/stop", stopStream)
	r.GET("/api/stream/status/:url", getStreamStatus)
	r.GET("/api/streams", listStreams)

	// 启动服务器
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
} 