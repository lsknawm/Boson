package controller

import (
	"Boson/dao"
	"Boson/model"
	"Boson/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GenerateQuiz(c *gin.Context) {
	var req model.QuizGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	questions, err := service.GenerateQuiz(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"quiz": questions, "count": len(questions)})
}

func ValidateAnswer(c *gin.Context) {
	var req model.UserAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	correct, stdAns, explain, err := service.ValidateAnswer(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"correct":     correct,
		"user_answer": req.UserAnswer,
		"std_answer":  stdAns,
		"explanation": explain,
	})
}

func GetExplanation(c *gin.Context) {
	q, err := dao.FindQuestionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
		return
	}
	if q.Validation != nil {
		c.JSON(http.StatusOK, q.Validation.Explanation)
	} else {
		c.JSON(http.StatusOK, gin.H{"text": "暂无解析"})
	}
}
