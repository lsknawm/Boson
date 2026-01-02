package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 配置信息
const (
	Host     = "43.142.0.45"
	Port     = "27017"
	User     = "admin"
	Password = "000000"
)

func main() {
	address := fmt.Sprintf("%s:%s", Host, Port)
	fmt.Printf("========================================\n")
	fmt.Printf("开始诊断 MongoDB 连接: %s\n", address)
	fmt.Printf("========================================\n\n")

	// --- 步骤 1: 基础 TCP 握手测试 ---
	fmt.Println("[步骤 1] 测试 TCP 端口连通性...")
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		fmt.Printf("❌ TCP 连接失败: %v\n", err)
		fmt.Println("   -> 原因推断: 云服务器安全组未放行，或 IP 错误，或服务器防火墙拦截。")
		return
	}
	fmt.Println("✅ TCP 握手成功！(网络层是通的)")

	// --- 步骤 2: Socket 稳定性测试 ---
	fmt.Println("\n[步骤 2] 测试 Socket 连接稳定性 (检测是否秒断)...")
	// 尝试读取一个字节，看服务端是否立即关闭连接
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 1)
	_, err = conn.Read(buffer)

	if err != nil {
		// EOF 意味着服务端主动关闭了连接
		if err.Error() == "EOF" {
			fmt.Printf("❌ 检测到服务端立即断开连接 (EOF)！\n")
			fmt.Println("   -> 高概率原因: MongoDB 容器正在无限重启（崩溃）。")
			fmt.Println("   -> 检查建议: 您的服务器 CPU 可能不支持 AVX 指令集，请务必换用 mongo:4.4 镜像。")
			conn.Close()
			return
		}
		// 超时是正常的，因为我们没发数据，服务端在等我们
		fmt.Println("✅ Socket 连接保持稳定 (未立即断开)。")
	}
	conn.Close()

	// --- 步骤 3: 驱动连接测试 (直连模式) ---
	fmt.Println("\n[步骤 3] 使用 Go 驱动尝试认证登录...")

	// 构造强制直连 URI
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=admin&connect=direct", User, Password, Host, Port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Printf("❌ 驱动配置错误: %v\n", err)
		return
	}

	// Ping 测试
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Printf("❌ 认证/Ping 失败: %v\n", err)
		fmt.Println("   -> 原因推断: 账号密码错误，或 authSource 不对，或网络不稳定。")
	} else {
		fmt.Println("🎉 恭喜！MongoDB 连接完全正常！")
		fmt.Println("   -> 结论: 之前的代码可能 URI 写错了，请使用上面的 URI 格式。")
	}
}
