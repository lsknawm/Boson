package handler

import (
	"Boson/model"
	"Boson/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInfoHandler 获取学科和章节列表
func GetInfoHandler(c *gin.Context) {
	data, err := service.GetQuizOptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code: 500,
			Msg:  "获取基础信息失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

// GenerateQuizHandler 生成试题
func GenerateQuizHandler(c *gin.Context) {
	var req model.GenerateQuizRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: nil,
		})
		return
	}

	questions, err := service.GetQuestions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code: 500,
			Msg:  "获取题目失败: " + err.Error(),
			Data: nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 200,
		Msg:  "success",
		Data: questions,
	})
}
