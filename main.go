package main

import (
	"context"
	"log"
	"net/http"
	"reflect"
	"time"

	"Boson/conf"
	"Boson/model"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	// 1. 初始化数据库连接
	initDB()

	// 2. 初始化 Gin 引擎
	r := gin.Default()

	// 跨域处理 (简单版)
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

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 题目管理接口
		questions := v1.Group("/questions")
		{
			questions.GET("", listQuestions)
			questions.GET("/:id", getQuestion)
			questions.POST("", createQuestion)
			questions.PUT("/:id", updateQuestion)
			questions.DELETE("/:id", deleteQuestion)
		}

		// 刷题与考试接口
		quiz := v1.Group("/quiz")
		{
			quiz.POST("/generate", generateQuiz)
			quiz.POST("/validate", validateAnswer)
			quiz.GET("/:id/explanation", getExplanation)
		}

		// 元数据接口
		meta := v1.Group("/meta")
		{
			meta.GET("/subjects", getSubjects)
			meta.GET("/tags", getTags)
		}
	}

	log.Printf("服务启动中，监听端口 %s", conf.ServerPort)
	r.Run(conf.ServerPort)
}

// ---------------- Database Init ----------------

func initDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoURI))
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("无法 Ping 通数据库: %v", err)
	}

	collection = client.Database(conf.DbName).Collection(conf.CollectionName)
	log.Println("MongoDB 连接成功！")
}

// ---------------- Handlers ----------------

// listQuestions 获取题目列表 (支持分页和筛选)
func listQuestions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if subject := c.Query("subject"); subject != "" {
		filter["subject"] = subject
	}
	if qType := c.Query("type"); qType != "" {
		filter["type"] = qType
	}
	if diff := c.Query("difficulty"); diff != "" {
		filter["meta.difficulty"] = diff
	}

	// 分页
	// 这里简化处理，实际可以使用 strconv.Atoi 转换 page/size
	// findOptions := options.Find()
	// findOptions.SetLimit(int64(pageSize))
	// findOptions.SetSkip(int64((page - 1) * pageSize))

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer cursor.Close(ctx)

	var questions []model.Question
	if err = cursor.All(ctx, &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据解析失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": questions, "count": len(questions)})
}

// getQuestion 获取单题详情
func getQuestion(c *gin.Context) {
	id := c.Param("id")
	q, err := findQuestionHelper(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到该题目"})
		return
	}
	c.JSON(http.StatusOK, q)
}

// createQuestion 创建题目
func createQuestion(c *gin.Context) {
	var q model.Question
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 强制清空 UUID，由 Mongo 生成
	q.UUID = primitive.NilObjectID

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "插入失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID, "message": "创建成功"})
}

// updateQuestion 更新题目
func updateQuestion(c *gin.Context) {
	id := c.Param("id")
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 不允许直接更新 _id
	delete(updateData, "uuid")
	delete(updateData, "_id")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	filter, _ := buildIdFilter(id)
	update := bson.M{"$set": updateData}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// deleteQuestion 删除题目
func deleteQuestion(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	filter, _ := buildIdFilter(id)
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

// generateQuiz 随机生成试卷
func generateQuiz(c *gin.Context) {
	var req model.QuizGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if req.Count <= 0 {
		req.Count = 5 // 默认 5 题
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	// 聚合管道: Match -> Sample
	matchStage := bson.D{{Key: "$match", Value: bson.M{"subject": req.Subject}}}
	if req.Difficulty != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.M{"subject": req.Subject, "meta.difficulty": req.Difficulty}}}
	}
	sampleStage := bson.D{{Key: "$sample", Value: bson.M{"size": req.Count}}}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, sampleStage})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成试卷失败"})
		return
	}
	defer cursor.Close(ctx)

	var questions []model.Question
	if err = cursor.All(ctx, &questions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析失败"})
		return
	}

	// *** 关键 ***: 返回给前端做题时，必须把答案 (Validation) 隐藏掉
	for i := range questions {
		questions[i].Validation = nil
	}

	c.JSON(http.StatusOK, gin.H{"quiz": questions, "count": len(questions)})
}

// validateAnswer 校验答案
func validateAnswer(c *gin.Context) {
	var req model.UserAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查出原始题目（包含答案）
	q, err := findQuestionHelper(req.QuestionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	if q.Validation == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "该题目缺少标准答案配置"})
		return
	}

	// 比对答案
	correct := compareAnswers(q.Type, q.Validation.Answer, req.UserAnswer)

	c.JSON(http.StatusOK, gin.H{
		"correct":     correct,
		"user_answer": req.UserAnswer,
		"std_answer":  q.Validation.Answer, // 判分后可以返回标准答案
		"explanation": q.Validation.Explanation,
	})
}

// getExplanation 单独获取解析
func getExplanation(c *gin.Context) {
	id := c.Param("id")
	q, err := findQuestionHelper(id)
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

// getSubjects 获取所有科目
func getSubjects(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	subjects, err := collection.Distinct(ctx, "subject", bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"subjects": subjects})
}

// getTags 获取所有标签
func getTags(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	tags, err := collection.Distinct(ctx, "meta.tags", bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// ---------------- Helpers ----------------

// findQuestionHelper 根据 hex UUID 或 自定义 ID (Q_xxx) 查找题目
func findQuestionHelper(id string) (*model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter, err := buildIdFilter(id)
	if err != nil {
		return nil, err
	}

	var q model.Question
	err = collection.FindOne(ctx, filter).Decode(&q)
	return &q, err
}

// buildIdFilter 尝试构建 ObjectID 过滤器，如果失败则回退到普通 ID 字符串
func buildIdFilter(id string) (bson.M, error) {
	// 尝试解析为 ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		return bson.M{"_id": objID}, nil
	}
	// 解析失败，说明可能是自定义ID (如 "Q_101")
	return bson.M{"id": id}, nil
}

// compareAnswers 简单的答案比对逻辑
func compareAnswers(qType string, std interface{}, user interface{}) bool {
	// 1. JSON 数字通常被解析为 float64，这里做简单处理
	// 实际生产中建议使用 deepEqual 或专门的比较库
	return reflect.DeepEqual(std, user)
}
