package main

import (
	"fmt"
	"net/http"
	"syscall"

)

// 定义Windows API函数和常量
var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procSendMessage  = user32.NewProc("SendMessageW")
	HWND_BROADCAST   = uintptr(0xFFFF)
	WM_SYSCOMMAND    = uintptr(0x0112)
	SC_MONITORPOWER  = uintptr(0xF170)
	MONITOR_OFF      = uintptr(2)
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

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Starting server on :8233")
	err := http.ListenAndServe(":8233", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
