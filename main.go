package main

import (
	"Boson/database"
	"Boson/router"
	"fmt"
)

func main() {
	// 1. 初始化数据库
	database.InitDB()

	// 2. 设置路由并启动服务
	r := router.SetupRouter()

	// 3. 启动服务
	fmt.Println("Server is running on port 8080")
	r.Run(":8080")
}
