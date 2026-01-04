package controller

import (
	"Boson/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSubjects(c *gin.Context) {
	subjects, err := dao.GetDistinctValues("subject")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subjects": subjects})
}

func GetTags(c *gin.Context) {
	tags, err := dao.GetDistinctValues("meta.tags")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
