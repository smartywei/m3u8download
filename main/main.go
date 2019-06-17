package main

import (
	"M3u8Tool/servers"
	"M3u8Tool/config"
)

func main() {
	start := make(chan bool, 1)
	//加载配置文件
	config.InitConfig()
	go servers.StartHttpServer() //启动HTTP服务
	go servers.StartDownloadServer() //启动下载服务
	go servers.StartTransFormServer() //启动转换服务
	<-start
}
