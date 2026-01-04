package dao

import (
	"Boson/conf"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Collection *mongo.Collection

// InitDB 初始化数据库连接
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoURI))
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("无法 Ping 通数据库: %v", err)
	}

	Collection = client.Database(conf.DbName).Collection(conf.CollectionName)
	log.Println("MongoDB 连接成功！")
}
