package service

import (
	"Boson/database"
	"Boson/model"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetQuizOptions 获取所有学科及对应的章节信息 (聚合查询)
func GetQuizOptions() ([]model.SubjectInfo, error) {
	collection := database.GetCollection()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 聚合管道
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
	collection := database.GetCollection()

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

	// 防止返回 nil
	if questions == nil {
		questions = []model.Question{}
	}

	return questions, nil
}
