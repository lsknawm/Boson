package database

import (
	"Boson/conf"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// InitDB 初始化数据库
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("正在连接 MongoDB...")
	clientOpts := options.Client().ApplyURI(conf.MongoURI)
	var err error
	Client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("连接配置错误: %v", err)
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("无法连接到 MongoDB: %v", err)
	}
	fmt.Println("MongoDB 连接成功！")
}

// GetCollection 获取默认集合的辅助函数
func GetCollection() *mongo.Collection {
	return Client.Database(conf.DBName).Collection(conf.CollName)
}
