package tools

import (
	"Boson/model"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const (
	MongoURI = "mongodb://localhost:27017"
	DBName   = "quanta_db"
	CollName = "data"
)

// InitDB 初始化数据库
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

// GetQuizOptions 获取所有学科及对应的章节信息 (聚合查询)
func GetQuizOptions() ([]model.SubjectInfo, error) {
	collection := client.Database(DBName).Collection(CollName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 聚合管道：
	// 1. $group: 按 subject 分组 (_id = "$subject")
	// 2. $addToSet: 将 meta.chapter 加入 chapters 数组 (自动去重)
	// 3. $sort: 按学科名排序
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$subject"},
			{Key: "chapters", Value: bson.D{{Key: "$addToSet", Value: "$meta.chapter"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []model.SubjectInfo
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// GetQuestions 根据筛选条件获取题目
func GetQuestions(req model.GenerateQuizRequest) ([]model.Question, error) {
	collection := client.Database(DBName).Collection(CollName)

	filter := bson.M{"subject": req.Subject}

	// 难度筛选
	if req.DifficultyStart != "" || req.DifficultyEnd != "" {
		start := req.DifficultyStart
		end := req.DifficultyEnd
		if start == "" {
			start = end
		}
		if end == "" {
			end = start
		}
		if start > end {
			start, end = end, start
		}
		filter["meta.difficulty"] = bson.M{"$gte": start, "$lte": end}
	}

	// 章节筛选 (正则匹配)
	if len(req.Chapters) > 0 {
		var orConditions []bson.M
		for _, kw := range req.Chapters {
			if kw == "" {
				continue
			}
			safePattern := regexp.QuoteMeta(kw)
			orConditions = append(orConditions, bson.M{
				"meta.chapter": primitive.Regex{Pattern: safePattern, Options: "i"},
			})
		}
		if len(orConditions) > 0 {
			filter["$or"] = orConditions
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

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

	// 防止返回 nil (虽然 gin 会处理为 null，但空数组对前端更友好)
	if questions == nil {
		questions = []model.Question{}
	}

	return questions, nil
}
