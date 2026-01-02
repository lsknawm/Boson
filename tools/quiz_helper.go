package tools

import (
	"Boson/model"
	"context"
	"fmt"
	"log"
	"regexp" // 用于转义正则特殊字符
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive" // 用于构建 MongoDB 正则对象
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

	// 1. 构建基础筛选条件 (Subject 是必须的)
	filter := bson.M{"subject": req.Subject}

	// 2. 难度范围筛选 (支持 A-C 这种区间)
	if req.DifficultyStart != "" || req.DifficultyEnd != "" {
		start := req.DifficultyStart
		end := req.DifficultyEnd

		// 默认值处理
		if start == "" {
			start = end
		}
		if end == "" {
			end = start
		}

		// 容错：保证 start <= end
		if start > end {
			start, end = end, start
		}

		filter["meta.difficulty"] = bson.M{
			"$gte": start,
			"$lte": end,
		}
	}

	// 3. 章节筛选 (改为模糊匹配)
	// 逻辑：只要题目章节名称中包含 request 中的任意一个关键词，即视为匹配
	if len(req.Chapters) > 0 {
		var orConditions []bson.M
		for _, kw := range req.Chapters {
			// QuoteMeta 确保关键词中的特殊符号（如 +, ?, *）被当作普通字符处理
			// Options: "i" 表示不区分大小写 (Case Insensitive)
			safePattern := regexp.QuoteMeta(kw)
			orConditions = append(orConditions, bson.M{
				"meta.chapter": primitive.Regex{Pattern: safePattern, Options: "i"},
			})
		}
		// 使用 $or 组合所有关键词条件
		filter["$or"] = orConditions
	}

	// 4. 题目数量限制
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// 5. 聚合管道：匹配 -> 随机抽样
	pipeline := []bson.M{
		{"$match": filter},                 // 筛选
		{"$sample": bson.M{"size": limit}}, // 随机乱序并截取
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
