package model

// 对应 questions.json 的核心结构 (保持不变)
type Question struct {
	ID         string                 `json:"id" bson:"id"`
	UUID       string                 `json:"uuid" bson:"uuid"`
	Type       string                 `json:"type" bson:"type"`
	Subject    string                 `json:"subject" bson:"subject"`
	Meta       QuestionMeta           `json:"meta" bson:"meta"`
	Content    QuestionContent        `json:"content" bson:"content"`
	Structure  map[string]interface{} `json:"structure" bson:"structure"`
	Validation map[string]interface{} `json:"validation" bson:"validation"`
}

type QuestionMeta struct {
	Chapter    string      `json:"chapter" bson:"chapter"`
	Difficulty string      `json:"difficulty" bson:"difficulty"`
	Score      interface{} `json:"score" bson:"score"`
}

type QuestionContent struct {
	Text  string `json:"text" bson:"text"`
	Image string `json:"image,omitempty" bson:"image,omitempty"`
	Code  string `json:"code,omitempty" bson:"code,omitempty"`
}

// 首页筛选请求参数 (已修改)
type GenerateQuizRequest struct {
	Subject         string   `json:"subject" binding:"required"` // 必选
	Chapters        []string `json:"chapters"`                   // 可选
	DifficultyStart string   `json:"difficulty_start"`           // 范围起点 (如 "A")
	DifficultyEnd   string   `json:"difficulty_end"`             // 范围终点 (如 "C")
	Limit           int      `json:"limit"`                      // 默认 10
}

// 统一 API 响应格式
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
