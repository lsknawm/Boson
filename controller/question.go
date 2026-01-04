package controller

import (
	"Boson/dao"
	"Boson/model"
	"Boson/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListQuestions Handler
func ListQuestions(c *gin.Context) {
	list, err := service.GetQuestionList(
		c.Query("subject"),
		c.Query("type"),
		c.Query("difficulty"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "count": len(list)})
}

// GetQuestion Handler
func GetQuestion(c *gin.Context) {
	q, err := dao.FindQuestionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该题目"})
		return
	}
	c.JSON(http.StatusOK, q)
}

// CreateQuestion Handler
func CreateQuestion(c *gin.Context) {
	var q model.Question
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := dao.CreateQuestion(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "创建成功"})
}

// UpdateQuestion Handler
func UpdateQuestion(c *gin.Context) {
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := dao.UpdateQuestion(c.Param("id"), updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteQuestion Handler
func DeleteQuestion(c *gin.Context) {
	if err := dao.DeleteQuestion(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
