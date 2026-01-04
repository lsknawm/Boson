package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	// 假设你会在其他包初始化数据库连接，这里仅做路由演示
)

func main() {
	// 初始化 Gin 引擎
	r := gin.Default()

	// 简单的健康检查
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong", "system": "Boson Quiz System"})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// ==========================
		// 1. 题目管理 (Questions CRUD)
		// ==========================
		questions := v1.Group("/questions")
		{
			questions.GET("", listQuestions)         // 获取列表 (带分页/筛选)
			questions.GET("/:id", getQuestion)       // 获取详情
			questions.POST("", createQuestion)       // 创建题目
			questions.PUT("/:id", updateQuestion)    // 更新题目
			questions.DELETE("/:id", deleteQuestion) // 删除题目
		}

		// ==========================
		// 2. 刷题与考试 (Quiz)
		// ==========================
		quiz := v1.Group("/quiz")
		{
			quiz.POST("/generate", generateQuiz)   // 生成一套题
			quiz.POST("/validate", validateAnswer) // 单题判分
		}

		// ==========================
		// 3. 元数据 (Metadata)
		// ==========================
		meta := v1.Group("/meta")
		{
			meta.GET("/subjects", getSubjects) // 获取科目列表
			meta.GET("/tags", getTags)         // 获取标签云
		}
	}

	// 启动服务 (默认 8080)
	r.Run(":8080")
}

// ---------------------------------------------------------
// Handler 占位符 (后续你需要在这里填入连接 MongoDB 的逻辑)
// ---------------------------------------------------------

func listQuestions(c *gin.Context) {
	// 获取 query 参数: c.Query("page"), c.Query("subject")
	// TODO: 查询 MongoDB 集合
	c.JSON(http.StatusOK, gin.H{"status": "todo", "action": "list_questions"})
}

func getQuestion(c *gin.Context) {
	id := c.Param("id")
	// TODO: 根据 ID 或 UUID 查找
	c.JSON(http.StatusOK, gin.H{"id": id, "action": "get_detail"})
}

func createQuestion(c *gin.Context) {
	// TODO: c.BindJSON(&question) -> InsertOne
	c.JSON(http.StatusCreated, gin.H{"action": "create"})
}

func updateQuestion(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id, "action": "update"})
}

func deleteQuestion(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"id": id, "action": "delete"})
}

func generateQuiz(c *gin.Context) {
	// 比如接收: { "subject": "Math", "size": 10 }
	// TODO: MongoDB Aggregate $sample 随机抽取
	c.JSON(http.StatusOK, gin.H{"action": "generate_random_quiz"})
}

func validateAnswer(c *gin.Context) {
	// 比如接收: { "id": "Q_001", "answer": "OPT_A" }
	// TODO: 对比数据库中的 validation.answer
	c.JSON(http.StatusOK, gin.H{"correct": true, "explanation": "..."})
}

func getSubjects(c *gin.Context) {
	// TODO: db.questions.distinct("subject")
	c.JSON(http.StatusOK, gin.H{"subjects": []string{"计算机科学", "English"}})
}

func getTags(c *gin.Context) {
	// TODO: db.questions.distinct("meta.tags")
	c.JSON(http.StatusOK, gin.H{"tags": []string{"HTTP", "算法", "语法"}})
}
