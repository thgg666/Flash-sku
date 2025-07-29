package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flashsku/seckill/internal/app"
)

// main 函数 - Go秒杀服务的入口点
// Main function - Entry point for Go seckill service
func main() {
	// 创建应用程序实例
	// Create application instance
	application, err := app.New()
	if err != nil {
		log.Fatalf("❌ 应用程序初始化失败: %v", err)
	}

	// 启动服务器的goroutine
	// Goroutine to start server
	go func() {
		log.Printf("🚀 秒杀服务启动中...")
		log.Printf("🚀 Seckill service starting...")
		if err := application.Start(); err != nil {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")
	log.Println("🛑 Shutting down server...")

	// 5秒超时的优雅关闭
	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Stop(ctx); err != nil {
		log.Fatalf("❌ 服务器强制关闭: %v", err)
	}

	log.Println("✅ 服务器已退出")
	log.Println("✅ Server exited")
}
