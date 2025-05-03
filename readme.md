# 🔌 closeMonitorByHttp
[![Golang 1.23.1](https://img.shields.io/badge/Go-1.23.1-blue.svg?style=flat)](https://golang.google.cn/doc/devel/release#go1.23.1)
[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-orange.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Release: v1.3](https://img.shields.io/badge/Release-v1.3-brightgreen.svg)](https://github.com/Billyas/closeMonitorByHttp/releases/tag/v1.3)

Windows显示器远程关闭工具，通过HTTP接口实现显示器电源管理。

## ✨ 功能特性
- 基于Windows API实现显示器关闭控制
- 提供简易HTTP接口（GET /）
- 支持系统托盘图标操作
- 自动生成可执行文件图标

## 🛠️ 编译安装
```bash
# 克隆仓库
git clone https://github.com/Billyas/closeMonitorByHttp.git
cd closeMonitorByHttp

# 编译项目
go build -ldflags -H=windowsgui

```

## 🚀 快速使用
```bash
# 启动服务（默认端口8233）
./closeMonitorByHttp.exe

# 发送关闭显示器请求
curl http://localhost:8233/
```



## 🔧 图标定制
1. 准备256x256像素PNG图片
2. 访问 [在线图标转换工具](https://www.icoconverter.com)
3. 转换后保存为icon.png到项目根目录
4. 执行编译项目重新生成可执行程序

