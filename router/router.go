package router

import (
	"Boson/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化 Gin 引擎并注册路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 配置 CORS
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

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api := r.Group("/api")
	{
		// 获取学科与章节信息 (前端初始化调用)
		api.GET("/info", handler.GetInfoHandler)
		// 生成试题 (点击开始刷题调用)
		api.POST("/generate-quiz", handler.GenerateQuizHandler)
	}

	return r
}
