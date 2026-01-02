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

// Config 配置区域
const (
	MongoURI     = "mongodb://localhost:27017"
	DBName       = "quanta_db"
	CollName     = "data"
	JsonFilePath = "./questions.json" // 确保路径正确
)

func main() {
	// 1. 初始化上下文
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 2. 连接数据库
	client, err := connectDB(ctx, MongoURI)
	if err != nil {
		log.Fatalf("[Error] 数据库连接失败: %v", err)
	}
	defer client.Disconnect(ctx)

	// 3. 读取本地文件
	questions, err := loadQuestions(JsonFilePath)
	if err != nil {
		log.Fatalf("[Error] 读取文件失败: %v", err)
	}
	if len(questions) == 0 {
		fmt.Println("[Info] 文件为空，无需处理。")
		return
	}

	// 4. 执行核心同步逻辑
	collection := client.Database(DBName).Collection(CollName)
	dirty, err := syncQuestions(ctx, collection, questions)
	if err != nil {
		log.Printf("[Error] 同步过程中发生错误: %v", err)
	}

	// 5. 如果有 UUID 变动，回写文件
	if dirty {
		if err := saveFileIfNeeded(JsonFilePath, questions); err != nil {
			log.Fatalf("[Error] 回写文件失败: %v", err)
		}
	}

	// 6. 维护索引
	ensureIndexes(ctx, collection)
}

// -------------------------------------------------------------------------
// 功能函数封装
// -------------------------------------------------------------------------

// connectDB 建立数据库连接
func connectDB(ctx context.Context, uri string) (*mongo.Client, error) {
	fmt.Println("正在连接 MongoDB...")
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	fmt.Println("MongoDB 连接成功！")
	return client, nil
}

// loadQuestions 读取并解析 JSON 文件为动态 Map
func loadQuestions(path string) ([]map[string]interface{}, error) {
	fmt.Printf("正在读取 %s ...\n", path)

	// 1. 打开文件流
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// 务必记得关闭文件，防止资源泄露
	defer file.Close()

	var questions []map[string]interface{}

	// 2. 创建解码器
	decoder := json.NewDecoder(file)

	// 【可选】防止将数字解析为 float64 (例如 id: 1234567890123456789 可能会精度丢失)
	// 开启后，数字会被解析为 json.Number 类型，你可以按需转为 Int64 或 Float64
	decoder.UseNumber()

	// 3. 执行解码
	if err := decoder.Decode(&questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// syncQuestions 核心同步逻辑
// 返回 bool 表示是否修改了内存中的 questions (需要回写文件)
func syncQuestions(ctx context.Context, coll *mongo.Collection, questions []map[string]interface{}) (bool, error) {
	isFileDirty := false
	fmt.Printf("开始处理 %d 道题目...\n", len(questions))

	for i, q := range questions {
		// 1. 清理数据: 移除 _id 防止不可变错误
		delete(q, "_id")

		// 2. 提取关键标识
		idVal, _ := q["id"].(string)
		uuidVal, _ := q["uuid"].(string)

		// 3. 查找数据库中是否已存在 (通过 UUID 或 ID)
		targetID, matchType := findExistingDoc(ctx, coll, uuidVal, idVal)

		if targetID != nil {
			// === 更新逻辑 (UPDATE) ===
			// 使用 $set 覆盖所有字段。因为 q 是 map[string]interface{}，
			// 所以无论 JSON 里加了什么新字段，这里都会自动同步进去。
			fmt.Printf("  [~] 更新 (ID: %-12s) | 匹配: %-4s\n", idVal, matchType)
			_, err := coll.UpdateOne(ctx, bson.M{"_id": targetID}, bson.M{"$set": q})
			if err != nil {
				log.Printf("      更新失败: %v", err)
			}
		} else {
			// === 插入逻辑 (INSERT) ===
			fmt.Printf("  [+] 插入 (ID: %-12s)\n", idVal)

			// 如果本地没有 UUID，插入后需要生成
			if uuidVal == "" {
				res, err := coll.InsertOne(ctx, q)
				if err != nil {
					log.Printf("      插入失败: %v", err)
					continue
				}
				// 补全 UUID
				if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
					newUUID := oid.Hex()

					// 更新内存 (用于回写文件)
					questions[i]["uuid"] = newUUID

					// 更新数据库 (补上 uuid 字段)
					coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"uuid": newUUID}})

					isFileDirty = true
					fmt.Printf("      -> 生成新 UUID: %s\n", newUUID)
				}
			} else {
				// 本地已有 UUID，直接插入
				_, err := coll.InsertOne(ctx, q)
				if err != nil {
					log.Printf("      插入失败: %v", err)
				}
			}
		}
	}
	return isFileDirty, nil
}

// findExistingDoc 尝试找到存在的文档
// 优先级: 1. UUID匹配  2. ID匹配
func findExistingDoc(ctx context.Context, coll *mongo.Collection, uuid string, id string) (interface{}, string) {
	var doc bson.M

	// 策略 A: UUID 匹配 (最强指纹)
	if uuid != "" {
		err := coll.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&doc)
		if err == nil {
			return doc["_id"], "UUID"
		}
	}

	// 策略 B: ID 匹配 (修复旧数据/兜底)
	if id != "" {
		err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&doc)
		if err == nil {
			return doc["_id"], "ID"
		}
	}

	return nil, ""
}

// saveFileIfNeeded 将内存数据回写到 JSON
func saveFileIfNeeded(path string, data []map[string]interface{}) error {
	fmt.Println("正在回写本地文件 (更新 UUID)...")
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

// ensureIndexes 确保必要的索引存在
func ensureIndexes(ctx context.Context, coll *mongo.Collection) {
	// 唯一索引: uuid
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	// 普通索引: id (用于快速查找)
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "id", Value: 1}},
	})
	// 普通索引: subject (用于前端筛选)
	coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "subject", Value: 1}},
	})

	fmt.Println("索引维护完成。")
}
