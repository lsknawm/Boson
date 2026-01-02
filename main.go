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

// Config 数据库配置
const (
	MongoURI = "mongodb://localhost:27017"
	DBName   = "quanta_db" // 目标数据库名
	CollName = "data"      // 目标集合(表)名
	JsonFile = "./script/questions.json"
)

func main() {
	// 1. 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 2. 连接 MongoDB
	fmt.Println("正在连接 MongoDB...")
	clientOpts := options.Client().ApplyURI(MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("连接配置错误: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("无法连接到 MongoDB: %v", err)
	}
	fmt.Println("MongoDB 连接成功！")

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatalf("断开连接失败: %v", err)
		}
	}()

	// 3. 读取本地 JSON
	fmt.Printf("正在读取 %s ...\n", JsonFile)
	fileBytes, err := os.ReadFile(JsonFile)
	if err != nil {
		log.Fatalf("无法读取文件: %v", err)
	}

	var questions []map[string]interface{}
	err = json.Unmarshal(fileBytes, &questions)
	if err != nil {
		log.Fatalf("JSON 解析失败: %v", err)
	}

	if len(questions) == 0 {
		fmt.Println("JSON 文件为空，无需处理。")
		return
	}

	// 4. 处理数据同步 (核心修改部分)
	collection := client.Database(DBName).Collection(CollName)
	isFileDirty := false // 标记本地文件是否需要更新

	fmt.Println("正在同步数据...")

	for i, q := range questions {
		// 获取当前的 uuid
		uuidVal, _ := q["uuid"].(string)

		// Case A: 新题 (JSON 中没有 uuid)
		if uuidVal == "" {
			fmt.Printf("  [+] 发现新题 (ID: %v)，正在插入并获取 MongoDB ID...\n", q["id"])

			// 1. 先插入数据 (MongoDB 会自动生成 _id)
			insertRes, err := collection.InsertOne(ctx, q)
			if err != nil {
				log.Printf("  [!] 插入失败: %v\n", err)
				continue
			}

			// 2. 获取生成的 _id (ObjectId)
			if oid, ok := insertRes.InsertedID.(primitive.ObjectID); ok {
				// 转为 16进制字符串
				generatedUUID := oid.Hex()

				// 3. 更新内存中的 JSON 数据
				questions[i]["uuid"] = generatedUUID

				// 4. 【重要】把这个字符串格式的 uuid 再更新回数据库的 uuid 字段
				// 这样数据库里的 uuid 字段和 _id 保持逻辑对应，且和 JSON 对应
				_, err := collection.UpdateOne(
					ctx,
					bson.M{"_id": oid}, // 根据刚才插入的主键查找
					bson.M{"$set": bson.M{"uuid": generatedUUID}},
				)
				if err != nil {
					log.Printf("  [!] 回写 UUID 到数据库字段失败: %v\n", err)
				}

				isFileDirty = true // 标记需要保存文件
			}

		} else {
			// Case B: 旧题 (JSON 中已有 uuid) -> 执行更新
			// 使用 uuid 字段作为过滤条件
			_, err := collection.UpdateOne(
				ctx,
				bson.M{"uuid": uuidVal},
				bson.M{"$set": q},                // 更新内容
				options.Update().SetUpsert(true), // 如果数据库里被误删了，会重新插入
			)
			if err != nil {
				log.Printf("  [!] 更新失败 (UUID: %s): %v\n", uuidVal, err)
			}
		}
	}

	// 5. 如果有新生成的 UUID，回写到本地 JSON 文件
	if isFileDirty {
		fmt.Println("检测到新生成的 UUID，正在回写本地文件...")

		// 格式化 JSON (带缩进)
		newFileBytes, err := json.MarshalIndent(questions, "", "  ")
		if err != nil {
			log.Fatalf("JSON 序列化失败: %v", err)
		}

		err = os.WriteFile(JsonFile, newFileBytes, 0644)
		if err != nil {
			log.Fatalf("无法写入文件: %v", err)
		}
		fmt.Println("本地 questions.json 已更新 (填充了 uuid)！")
	} else {
		fmt.Println("本地文件无需更新 (无新题添加)。")
	}

	// 6. 确保索引
	createIndex(ctx, collection)
}

func createIndex(ctx context.Context, coll *mongo.Collection) {
	// 为 uuid 字段创建唯一索引
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		// 忽略索引已存在的错误
	} else {
		fmt.Println("索引检查完成。")
	}
}
