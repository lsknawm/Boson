package model

// Question 核心题目结构
type Question struct {
	ID         string             `json:"id" bson:"id"`
	UUID       string             `json:"uuid" bson:"uuid"`
	Type       string             `json:"type" bson:"type"`       // single_choice, multiple_choice, true_false, fill_blank, short_answer
	Subject    string             `json:"subject" bson:"subject"` // 学科
	Meta       QuestionMeta       `json:"meta" bson:"meta"`
	Content    QuestionContent    `json:"content" bson:"content"`       // 题干
	Structure  QuestionStructure  `json:"structure" bson:"structure"`   // 题目结构
	Validation QuestionValidation `json:"validation" bson:"validation"` // 答案
}

// QuestionMeta 元数据
type QuestionMeta struct {
	Chapter    string      `json:"chapter" bson:"chapter"`       // 章节
	Difficulty string      `json:"difficulty" bson:"difficulty"` // 难度
	Score      interface{} `json:"score" bson:"score"`           // 分数 (可能是 int 或 float)
}

// QuestionContent 内容块 (用于题干、解析)
type QuestionContent struct {
	Text     string `json:"text" bson:"text"`
	Image    string `json:"image,omitempty" bson:"image,omitempty"` // Base64 或 URL
	Code     string `json:"code,omitempty" bson:"code,omitempty"`
	HasImage bool   `json:"has_image,omitempty" bson:"has_image,omitempty"`
	// 以下字段用于代码题运行结果，非必须可省略
	CodeError    bool   `json:"code_error,omitempty" bson:"code_error,omitempty"`
	CodeRunCount int    `json:"code_run_count,omitempty" bson:"code_run_count,omitempty"`
	DebugMsg     string `json:"debug_msg,omitempty" bson:"debug_msg,omitempty"`
}

// QuestionStructure 题目结构 (兼容所有题型)
type QuestionStructure struct {
	Layout  string   `json:"layout,omitempty" bson:"layout,omitempty"`   // 布局: vertical, horizontal
	Options []Option `json:"options,omitempty" bson:"options,omitempty"` // 用于: 单选/多选/判断
	Blanks  []Blank  `json:"blanks,omitempty" bson:"blanks,omitempty"`   // 用于: 填空
}

// Option 选项
type Option struct {
	ID       string `json:"id" bson:"id"` // 选项ID: A, B, T, F
	Text     string `json:"text" bson:"text"`
	Image    string `json:"image,omitempty" bson:"image,omitempty"`
	Code     string `json:"code,omitempty" bson:"code,omitempty"`
	HasImage bool   `json:"has_image,omitempty" bson:"has_image,omitempty"`
}

// Blank 填空位
type Blank struct {
	ID          string `json:"id" bson:"id"`                   // 填空ID: b1, b2
	Placeholder string `json:"placeholder" bson:"placeholder"` // 占位符提示
}

// QuestionValidation 答案验证
type QuestionValidation struct {
	// Answer 使用 interface{} 接收，因为可能是 string(单选), []string(多选), map(填空)
	Answer      interface{}     `json:"answer" bson:"answer"`
	Explanation QuestionContent `json:"explanation" bson:"explanation"` // 解析
}

// --- 请求与响应相关 ---

// GenerateQuizRequest 生成试题请求参数
type GenerateQuizRequest struct {
	Subject         string   `json:"subject" binding:"required"`
	Chapters        []string `json:"chapters"`
	DifficultyStart string   `json:"difficulty_start"`
	DifficultyEnd   string   `json:"difficulty_end"`
	Limit           int      `json:"limit"`
}

// SubjectInfo 学科信息 (用于前端下拉菜单)
type SubjectInfo struct {
	Name     string   `json:"name" bson:"_id"`          // 聚合查询时 _id 即为 subject
	Chapters []string `json:"chapters" bson:"chapters"` // 包含的章节列表
}

// Response 统一 API 响应
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
