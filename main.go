package main

import (
	"Boson/route"
	"Boson/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化数据库
	tools.InitDB()

	// 2. 初始化 Gin 引擎
	r := gin.Default()

	// 3. 配置 CORS (允许跨域请求，方便前端调试)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// 4. 注册路由
	route.RegisterRoutes(r)

	// 5. 启动服务 (默认 8080 端口)
	r.Run(":8080")
}
