package router

import (
	"Boson/controller"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "system": "Boson Quiz System"})
	})

	v1 := r.Group("/api/v1")
	{
		// 题目管理
		qGroup := v1.Group("/questions")
		{
			qGroup.GET("", controller.ListQuestions)
			qGroup.GET("/:id", controller.GetQuestion)
			qGroup.POST("", controller.CreateQuestion)
			qGroup.PUT("/:id", controller.UpdateQuestion)
			qGroup.DELETE("/:id", controller.DeleteQuestion)
		}

		// 刷题考试
		quizGroup := v1.Group("/quiz")
		{
			quizGroup.POST("/generate", controller.GenerateQuiz)
			quizGroup.POST("/validate", controller.ValidateAnswer)
			quizGroup.GET("/:id/explanation", controller.GetExplanation)
		}

		// 元数据
		metaGroup := v1.Group("/meta")
		{
			metaGroup.GET("/subjects", controller.GetSubjects)
			metaGroup.GET("/tags", controller.GetTags)
		}
	}
	return r
}
