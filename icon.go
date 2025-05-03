package main

import (
	"embed"
	"fmt"
)

//go:embed icon.ico
var iconFS embed.FS

// GetDefaultIconData 返回默认的图标数据
func GetDefaultIconData() ([]byte, error) {
	// 从嵌入的文件系统中读取图标数据
	iconData, err := iconFS.ReadFile("icon.ico")
	if err != nil {
		return nil, fmt.Errorf("无法读取图标文件: %v", err)
	}

	// 暂时移除图标数据验证逻辑，后续可根据需要添加

	// 可以在这里添加更多的图标验证逻辑，例如检查图标尺寸等

	return iconData, nil
}
