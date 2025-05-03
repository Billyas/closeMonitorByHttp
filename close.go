package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"net"
	"net/http"
	"os"
	"syscall"
)

// 定义Windows API函数和常量
var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procSendMessage = user32.NewProc("SendMessageW")
	HWND_BROADCAST  = uintptr(0xFFFF)
	WM_SYSCOMMAND   = uintptr(0x0112)
	SC_MONITORPOWER = uintptr(0xF170)
	MONITOR_OFF     = uintptr(2)
)

func turnOffMonitor() {
	// 调用SendMessage来关闭显示器
	ret, _, err := procSendMessage.Call(HWND_BROADCAST, WM_SYSCOMMAND, SC_MONITORPOWER, MONITOR_OFF)
	if ret == 0 {
		fmt.Println("Failed to turn off the monitor:", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// 获取客户端的 IP 地址
		clientIP := r.RemoteAddr
		// 打印请求关闭的 IP 地址
		fmt.Printf("Received request to turn off the screen from IP: %s\n", clientIP)
		fmt.Fprintln(w, "Screen will be turned off")
		go turnOffMonitor()
	} else {
		http.NotFound(w, r)
	}
}

// 启动HTTP服务器
func startServer() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server on :8233")
	err := http.ListenAndServe(":8233", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// 初始化系统托盘
// 获取本地 IPv4 地址
func getLocalIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		// 检查网络地址是否为 IPv4 地址
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ipv4 := ipNet.IP.To4()
		if ipv4 == nil || ipv4.IsLoopback() {
			continue
		}
		return ipv4.String(), nil
	}

	return "", fmt.Errorf("未找到本地 IPv4 地址")
}

func onReady() {
	iconData, err := GetDefaultIconData()
	if err != nil {
		fmt.Printf("警告：加载系统托盘图标失败: %v，程序将继续运行。\n", err)
	} else {
		// 设置托盘图标和提示文字
		systray.SetIcon(iconData)

	}
	systray.SetTitle("显示器控制")
	systray.SetTooltip("通过HTTP控制显示器")

	// 添加退出菜单项
	mQuit := systray.AddMenuItem("退出", "退出应用")

	// 启动HTTP服务
	go startServer()

	// 处理退出事件
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
		os.Exit(0)
	}()

	// 获取本地 IPv4 地址
	localIP, err := getLocalIPv4()
	if err != nil {
		fmt.Printf("获取本地 IPv4 地址失败: %v\n", err)
		localIP = "未知"
	}

	// 修改鼠标提示内容
	tooltip := fmt.Sprintf("通过 HTTP 控制显示器，服务地址: http://%s:8233", localIP)
	systray.SetTooltip(tooltip)
}

// 清理资源
func onExit() {
	// 在这里可以添加清理代码
}

func main() {
	// 启动系统托盘
	systray.Run(onReady, onExit)
}
