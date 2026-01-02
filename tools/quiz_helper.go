package tools

import (
	"Boson/model"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const (
	MongoURI = "mongodb://localhost:27017"
	DBName   = "quanta_db"
	CollName = "data"
)

// InitDB 初始化数据库连接
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("正在连接 MongoDB...")
	clientOpts := options.Client().ApplyURI(MongoURI)
	var err error
	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("连接配置错误: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("无法连接到 MongoDB: %v", err)
	}
	fmt.Println("MongoDB 连接成功！")
}

// GetQuestions 根据筛选条件从数据库随机获取题目
func GetQuestions(req model.GenerateQuizRequest) ([]model.Question, error) {
	collection := client.Database(DBName).Collection(CollName)

	// 1. 构建筛选条件 (Match)
	filter := bson.M{"subject": req.Subject}

	// 难度范围处理
	// 逻辑：如果是 "A" 到 "C"，则包含 A, B, C。
	if req.DifficultyStart != "" || req.DifficultyEnd != "" {
		start := req.DifficultyStart
		end := req.DifficultyEnd

		// 如果只传了一个值，默认 Start=End (相当于单选)
		if start == "" {
			start = end
		}
		if end == "" {
			end = start
		}

		// 容错处理：如果 Start > End (例如 Start="C", End="A")，交换它们以确保查询有效
		if start > end {
			start, end = end, start
		}

		filter["meta.difficulty"] = bson.M{
			"$gte": start,
			"$lte": end,
		}
	}

	// 章节筛选
	if len(req.Chapters) > 0 {
		filter["meta.chapter"] = bson.M{"$in": req.Chapters}
	}

	// 2. 设置题目数量限制
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// 3. 聚合管道：匹配 -> 随机抽样
	pipeline := []bson.M{
		{"$match": filter},
		{"$sample": bson.M{"size": limit}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []model.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	if questions == nil {
		questions = []model.Question{}
	}

	return questions, nil
}
