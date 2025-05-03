package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

// 定义Windows API函数和常量
var (
	user32          = syscall.NewLazyDLL("user32.dll")
	procSendMessage = user32.NewProc("SendMessageW")
	procMessageBox  = user32.NewProc("MessageBoxW")

	MB_OK              = uintptr(0x00000000)
	MB_ICONINFORMATION = uintptr(0x00000040)
	MB_YESNO           = uintptr(0x00000004)
	HWND_BROADCAST     = uintptr(0xFFFF)
	WM_SYSCOMMAND      = uintptr(0x0112)
	SC_MONITORPOWER    = uintptr(0xF170)
	MONITOR_OFF        = uintptr(2)
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

// setAutoStart 设置程序开机自启动
// checkAutoStart 检测当前是否设置开机启动
func checkAutoStart() (bool, error) {
	key, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return false, nil
	}
	defer key.Close()

	_, _, err = key.GetStringValue("CloseMonitorByHttp")
	return err == nil, nil
}

// setAutoStart 设置/取消开机自启动
func setAutoStart(enabled bool) error {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %v", err)
	}

	// 添加静默启动参数
	if !enabled {
		key, err := registry.OpenKey(
			registry.CURRENT_USER,
			`Software\Microsoft\Windows\CurrentVersion\Run`,
			registry.ALL_ACCESS,
		)
		if err != nil {
			return fmt.Errorf("打开注册表失败: %v", err)
		}
		defer key.Close()
		return key.DeleteValue("CloseMonitorByHttp")
	}

	command := fmt.Sprintf("\"%s\" -auto", exePath)

	// 打开注册表项
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf("创建注册表项失败: %v", err)
	}
	defer key.Close()

	// 设置注册表值
	if err := key.SetStringValue("CloseMonitorByHttp", command); err != nil {
		return fmt.Errorf("写入注册表失败: %v", err)
	}
	return nil
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

	// 添加带复选框的启动项
	mAutoStart := systray.AddMenuItemCheckbox("开机启动", "开机自动启动", false)
	// 初始化启动项状态（默认不勾选）
	if enabled, _ := checkAutoStart(); enabled {
		mAutoStart.Check()
	} else {
		mAutoStart.Uncheck()
	}

	// 处理启动项点击事件
	go func() {
		for {
			<-mAutoStart.ClickedCh
			currentState := mAutoStart.Checked()
			if err := setAutoStart(!currentState); err == nil {
				if !currentState {
					mAutoStart.Check()
				} else {
					mAutoStart.Uncheck()
				}
				systray.SetTooltip(fmt.Sprintf("当前启动状态: %t", !currentState))
			}
		}
	}()

	// 添加关于菜单项
	mAbout := systray.AddMenuItem("关于", "显示项目信息")
	// 添加退出菜单项
	mQuit := systray.AddMenuItem("退出", "退出应用")

	// 处理关于点击事件
	go func() {
		for {
			<-mAbout.ClickedCh
			showAboutDialog()
		}
	}()

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

// showAboutDialog 显示关于对话框
func showAboutDialog() {
	const title = "关于 closeMonitorByHttp"
	content := "项目名称: closeMonitorByHttp\n" +
		"GitHub: https://github.com/Billyas/closeMonitorByHttp\n" +
		"功能描述: 基于HTTP协议的远程显示器关闭工具\n\n是否要访问项目主页？"

	result, _, _ := procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(content))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(MB_YESNO|MB_ICONINFORMATION),
	)

	if result == 6 { // IDYES
		cmd := exec.Command("cmd", "/c", "start", "https://github.com/Billyas/closeMonitorByHttp")
		if err := cmd.Start(); err != nil {
			fmt.Printf("无法打开浏览器: %v\n", err)
		}
	}
}

// 清理资源
func onExit() {
	// 在这里可以添加清理代码
}

func main() {
	// 启动系统托盘
	systray.Run(onReady, onExit)
}
