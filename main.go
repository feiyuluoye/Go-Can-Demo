package main

import (
	"log"

	gocan "go-can.com/src"
)

func main() {
	// 创建mock数据
	if err := gocan.LoadMockData(); err != nil {
		log.Printf("Failed to create mock data: %v", err)
	}

	// 创建API服务器
	server := gocan.NewAPIServer()

	// 启动API服务器
	log.Println("Starting CAN API server on :8080")
	if err := server.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
