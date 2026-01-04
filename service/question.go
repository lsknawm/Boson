package service

import (
	"Boson/dao"
	"Boson/model"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

// GetQuestionList 获取处理后的列表
func GetQuestionList(subject, qType, difficulty string) ([]model.Question, error) {
	filter := bson.M{}
	if subject != "" {
		filter["subject"] = subject
	}
	if qType != "" {
		filter["type"] = qType
	}
	if difficulty != "" {
		filter["meta.difficulty"] = difficulty
	}
	return dao.ListQuestions(filter)
}

// GenerateQuiz 生成试卷 (业务逻辑：必须隐藏答案)
func GenerateQuiz(req model.QuizGenerateRequest) ([]model.Question, error) {
	if req.Count <= 0 {
		req.Count = 5
	}

	matchStage := bson.D{{Key: "$match", Value: bson.M{"subject": req.Subject}}}
	if req.Difficulty != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.M{"subject": req.Subject, "meta.difficulty": req.Difficulty}}}
	}
	sampleStage := bson.D{{Key: "$sample", Value: bson.M{"size": req.Count}}}

	questions, err := dao.AggregateQuestions(mongo.Pipeline{matchStage, sampleStage})
	if err != nil {
		return nil, err
	}

	// 核心业务：隐藏 Validation 字段
	for i := range questions {
		questions[i].Validation = nil
	}
	return questions, nil
}

// ValidateAnswer 判题逻辑
func ValidateAnswer(req model.UserAnswerRequest) (bool, interface{}, map[string]interface{}, error) {
	q, err := dao.FindQuestionByID(req.QuestionID)
	if err != nil {
		return false, nil, nil, errors.New("题目不存在")
	}

	if q.Validation == nil {
		return false, nil, nil, errors.New("该题目缺少标准答案配置")
	}

	// 业务逻辑：比对答案
	isCorrect := reflect.DeepEqual(q.Validation.Answer, req.UserAnswer)

	return isCorrect, q.Validation.Answer, q.Validation.Explanation, nil
}
