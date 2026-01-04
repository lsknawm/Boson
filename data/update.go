package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Configuration
const (
	MongoURI       = "mongodb://localhost:27017"
	DbName         = "quiz_system"
	CollectionName = "questions"
	FileName       = "questions.json"
)

// Question 结构体
type Question struct {
	// ID 是你的人类可读标识 (如 Q_001)，现在只是一个普通字段，允许修改
	ID string `json:"id" bson:"id"`

	// UUID 对应 MongoDB 的 _id。
	// bson:"-" 表示生成 BSON 数据准备存入数据库时，忽略这个字段。
	// 因为我们会手动把这个值转换为 _id 放在 Filter 中，或者让 Mongo 自动生成。
	// 这样可以避免数据库里同时出现 "_id" 和 "uuid" 两个字段。
	UUID string `json:"uuid" bson:"-"`

	Type       string                 `json:"type" bson:"type"`
	Subject    string                 `json:"subject" bson:"subject"`
	Meta       map[string]interface{} `json:"meta" bson:"meta"`
	Content    map[string]interface{} `json:"content" bson:"content"`
	Structure  map[string]interface{} `json:"structure" bson:"structure"`
	Validation map[string]interface{} `json:"validation" bson:"validation"`
}

func main() {
	// 1. 读取 JSON 文件
	log.Printf("正在读取文件: %s...", FileName)
	fileBytes, err := os.ReadFile(FileName)
	if err != nil {
		log.Fatalf("无法读取文件: %v", err)
	}

	var questions []Question
	if err := json.Unmarshal(fileBytes, &questions); err != nil {
		log.Fatalf("JSON 解析失败: %v", err)
	}

	// 2. 连接 MongoDB
	log.Println("正在连接 MongoDB...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(DbName).Collection(CollectionName)

	// 3. 遍历处理数据
	updatedCount := 0
	hasChanges := false

	for i := range questions {
		q := &questions[i]

		// === 分支 A: 新题目 (UUID 为空) ===
		if q.UUID == "" {
			// 插入操作：直接把结构体存进去 (bson:"-" 会忽略空的 UUID 字段)
			// MongoDB 会自动生成 _id
			result, err := collection.InsertOne(ctx, q)
			if err != nil {
				log.Printf("错误: 插入题目 [%s] 失败: %v", q.ID, err)
				continue
			}

			// 获取生成的 ObjectID 并回填到结构体
			if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
				newUUID := oid.Hex()
				q.UUID = newUUID
				hasChanges = true
				log.Printf("新题 [%s] 已插入，生成 UUID: %s", q.ID, newUUID)
			}
			updatedCount++

		} else {
			// === 分支 B: 已有题目 (UUID 有值) ===
			// 我们以 UUID (即 _id) 为绝对标准来寻找并更新数据

			// 1. 将字符串 UUID 转回 ObjectID
			objID, err := primitive.ObjectIDFromHex(q.UUID)
			if err != nil {
				log.Printf("警告: 题目 [%s] 的 UUID [%s] 格式错误，跳过同步。", q.ID, q.UUID)
				continue
			}

			// 2. 构造过滤器 ( WHERE _id = objID )
			filter := bson.M{"_id": objID}

			// 3. 构造更新内容
			// 使用 $set 将 JSON 里的最新内容覆盖到数据库
			// 注意：q 结构体里的 UUID 字段有 bson:"-" 标签，所以不会把 _id 覆盖写乱
			update := bson.M{"$set": q}

			// 4. 执行 Upsert
			// Upsert=true 的作用是：如果数据库里不小心删了这个 ID 的数据，
			// 这里会强制用这个 ID 重新创建一条，保证数据不丢失。
			opts := options.Update().SetUpsert(true)

			_, err = collection.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Printf("错误: 更新题目 [%s] (UUID: %s) 失败: %v", q.ID, q.UUID, err)
			} else {
				// 可以在这里打印日志，但为了不刷屏，仅计数
				updatedCount++
			}
		}
	}

	fmt.Printf("\n同步完成！共处理 %d 道题目。\n", updatedCount)

	// 4. 只有当生成了新 UUID 时，才回写 JSON 文件
	if hasChanges {
		log.Println("发现新生成的 UUID，正在回写到 questions.json ...")

		// 格式化 JSON (4空格缩进，美观)
		outputBytes, err := json.MarshalIndent(questions, "", "    ")
		if err != nil {
			log.Fatalf("无法序列化 JSON: %v", err)
		}

		if err := os.WriteFile(FileName, outputBytes, 0644); err != nil {
			log.Fatalf("无法写入文件: %v", err)
		}
		log.Println("文件回写成功！")
	} else {
		log.Println("没有新题目插入，无需修改本地文件。")
	}
}
