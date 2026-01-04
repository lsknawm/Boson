package dao

import (
	"Boson/conf"
	"Boson/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ListQuestions 查询题目列表
func ListQuestions(filter bson.M) ([]model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	cursor, err := Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []model.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

// FindQuestionByID 根据 ID 或 UUID 查找题目
func FindQuestionByID(id string) (*model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := buildIdFilter(id)
	var q model.Question
	err := Collection.FindOne(ctx, filter).Decode(&q)
	return &q, err
}

// CreateQuestion 创建题目
func CreateQuestion(q model.Question) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	// 确保 ID 由 Mongo 生成
	q.UUID = primitive.NilObjectID
	result, err := Collection.InsertOne(ctx, q)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

// UpdateQuestion 更新题目
func UpdateQuestion(id string, updateData map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	// 安全清理
	delete(updateData, "uuid")
	delete(updateData, "_id")

	filter := buildIdFilter(id)
	update := bson.M{"$set": updateData}

	_, err := Collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteQuestion 删除题目
func DeleteQuestion(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	filter := buildIdFilter(id)
	_, err := Collection.DeleteOne(ctx, filter)
	return err
}

// AggregateQuestions 聚合查询（用于随机组卷）
func AggregateQuestions(pipeline mongo.Pipeline) ([]model.Question, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()

	cursor, err := Collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []model.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}
	return questions, nil
}

// GetDistinctValues 获取唯一值 (用于元数据)
func GetDistinctValues(field string) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ContextTimeout)*time.Second)
	defer cancel()
	return Collection.Distinct(ctx, field, bson.M{})
}

// buildIdFilter 内部辅助函数
func buildIdFilter(id string) bson.M {
	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		return bson.M{"_id": objID}
	}
	return bson.M{"id": id}
}
