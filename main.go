package main

import (
	"Boson/conf"
	"Boson/dao"
	"Boson/router"
	"log"
)

func main() {
	// 1. 初始化数据库
	dao.InitDB()

	// 2. 初始化路由
	r := router.InitRouter()

	// 3. 启动服务
	log.Printf("服务启动中，监听端口 %s", conf.ServerPort)
	if err := r.Run(conf.ServerPort); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
