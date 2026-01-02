package route

import (
	"Boson/model"
	"Boson/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GenerateQuizHandler 处理生成试题的请求
func GenerateQuizHandler(c *gin.Context) {
	var req model.GenerateQuizRequest

	// 1. 绑定并校验 JSON 参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 2. 从数据库获取题目
	questions, err := tools.GetQuestions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code: 500,
			Msg:  "服务器内部错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	// 3. 返回成功响应
	c.JSON(http.StatusOK, model.Response{
		Code: 200,
		Msg:  "success",
		Data: questions,
	})
}

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 简单的健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api := r.Group("/api")
	{
		// 首页 -> 刷题页：提交配置，获取题目列表
		api.POST("/generate-quiz", GenerateQuizHandler)
	}
}
