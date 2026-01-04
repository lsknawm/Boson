package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Question 对应数据库中的题目结构
type Question struct {
	// UUID 映射 MongoDB 的 _id
	// omitempty 保证在插入新题时，如果为空，由 Mongo 自动生成
	UUID primitive.ObjectID `json:"uuid" bson:"_id,omitempty"`

	// ID 是人类可读的标识 (如 Q_001)
	ID string `json:"id" bson:"id"`

	Type       string                 `json:"type" bson:"type"`
	Subject    string                 `json:"subject" bson:"subject"`
	Meta       MetaInfo               `json:"meta" bson:"meta"`
	Content    map[string]interface{} `json:"content" bson:"content"`
	Structure  map[string]interface{} `json:"structure" bson:"structure"`
	Validation *ValidationInfo        `json:"validation,omitempty" bson:"validation,omitempty"`
}

// MetaInfo 元数据结构，方便类型安全的操作（可选，也可以用 map）
type MetaInfo struct {
	Chapter    []string `json:"chapter" bson:"chapter"`
	Tags       []string `json:"tags" bson:"tags"`
	Difficulty string   `json:"difficulty" bson:"difficulty"`
	Score      int      `json:"score" bson:"score"`
}

// ValidationInfo 验证信息结构
type ValidationInfo struct {
	Answer      interface{}            `json:"answer" bson:"answer"`
	Explanation map[string]interface{} `json:"explanation" bson:"explanation"`
}

// UserAnswerRequest 用于接收用户提交的答案进行校验
type UserAnswerRequest struct {
	QuestionID string      `json:"question_id"` // 支持 UUID 或 Q_id
	UserAnswer interface{} `json:"user_answer"`
}

// QuizGenerateRequest 生成试卷的请求参数
type QuizGenerateRequest struct {
	Subject    string `json:"subject"`
	Difficulty string `json:"difficulty"` // 可选
	Count      int    `json:"count"`      // 题目数量
}
