package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gocan "go-can.com/src"
)

// 自动测试MOCK数据和订阅功能
func main() {
	fmt.Println("=== CAN API 自动MOCK数据测试 ===")
	fmt.Println("用法: go run test_auto_mock.go [test|subscribe]")

	if len(os.Args) < 2 {
		fmt.Println("请指定测试模式: test 或 subscribe")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "test":
		// 加载mock数据
		fmt.Println("\n1. 加载MOCK数据...")
		if err := loadAndTestMockData(); err != nil {
			log.Fatalf("测试失败: %v", err)
		}
		fmt.Println("\n=== 自动测试完成 ===")

	case "subscribe":
		fmt.Println("\n启动SSE订阅测试...")
		testSubscribeMode()

	default:
		fmt.Println("未知模式，请使用 'test' 或 'subscribe'")
	}
}

func loadAndTestMockData() error {
	// 测试连接mock数据
	fmt.Println("\n2. 测试连接MOCK数据...")
	connectData, err := loadConnectMockData()
	if err != nil {
		return err
	}

	// 为每个连接通道创建会话
	var sessionIDs []string
	for _, conn := range connectData {
		sessionID, err := testConnect(conn.Channel)
		if err != nil {
			return err
		}
		sessionIDs = append(sessionIDs, sessionID)
		fmt.Printf("  创建会话: %s -> %s\n", conn.Channel, sessionID)
	}

	// 测试发送mock数据
	fmt.Println("\n3. 测试发送MOCK数据...")
	sendData, err := loadSendMockData()
	if err != nil {
		return err
	}

	for _, sendReq := range sendData {
		// 使用第一个会话ID
		sendReq.SessionID = sessionIDs[0]
		if err := testSend(sendReq); err != nil {
			return err
		}
		fmt.Printf("  发送CAN帧: ID=0x%X\n", sendReq.Frame.ID)
		time.Sleep(500 * time.Millisecond)
	}

	// 测试订阅mock数据
	fmt.Println("\n4. 测试订阅MOCK数据...")
	subscribeData, err := loadSubscribeMockData()
	if err != nil {
		return err
	}

	// 启动goroutine测试订阅
	for _, sub := range subscribeData {
		sub.SessionID = sessionIDs[0]
		go testSubscribe(sub)
	}

	// 等待一段时间让订阅测试运行
	fmt.Println("  订阅测试运行中...等待5秒")
	time.Sleep(5 * time.Second)

	// 测试断开连接
	fmt.Println("\n5. 测试断开连接...")
	for _, sessionID := range sessionIDs {
		if err := testDisconnect(sessionID); err != nil {
			return err
		}
		fmt.Printf("  断开会话: %s\n", sessionID)
	}

	return nil
}

func loadConnectMockData() ([]gocan.ConnectRequest, error) {
	data, err := os.ReadFile("data/connect_mock.json")
	if err != nil {
		return nil, err
	}

	var connectData []gocan.ConnectRequest
	err = json.Unmarshal(data, &connectData)
	if err != nil {
		return nil, err
	}

	return connectData, nil
}

func loadSendMockData() ([]gocan.SendRequest, error) {
	data, err := os.ReadFile("data/send_mock.json")
	if err != nil {
		return nil, err
	}

	var sendData []gocan.SendRequest
	err = json.Unmarshal(data, &sendData)
	if err != nil {
		return nil, err
	}

	return sendData, nil
}

func loadSubscribeMockData() ([]gocan.SubscribeRequest, error) {
	data, err := os.ReadFile("data/subscribe_mock.json")
	if err != nil {
		return nil, err
	}

	var subscribeData []gocan.SubscribeRequest
	err = json.Unmarshal(data, &subscribeData)
	if err != nil {
		return nil, err
	}

	return subscribeData, nil
}

func testConnect(channel string) (string, error) {
	reqBody := map[string]string{"channel": channel}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:8080/api/can/connect", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result gocan.ConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Status != "connected" {
		return "", fmt.Errorf("连接失败: %s", result.Message)
	}

	return result.ID, nil
}

func testSend(sendReq gocan.SendRequest) error {
	jsonData, _ := json.Marshal(sendReq)

	resp, err := http.Post("http://localhost:8080/api/can/send", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result gocan.SendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "sent" {
		return fmt.Errorf("发送失败: %s", result.Message)
	}

	return nil
}

func testSubscribe(sub gocan.SubscribeRequest) {
	// 使用SSE测试订阅 - GET请求带查询参数
	url := fmt.Sprintf("http://localhost:8080/api/can/subscribe?id=%s&canId=%d", sub.SessionID, sub.CanID)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("订阅失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("订阅请求失败: %s", resp.Status)
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("订阅连接关闭: %v", err)
			break
		}
		if strings.HasPrefix(line, "data: ") {
			fmt.Printf("  收到订阅消息: %s", line[6:])
		}
	}
}

func testSubscribeMode() {
	// 创建会话用于订阅测试
	sessionID, err := testConnect("PCAN_USBBUS1")
	if err != nil {
		log.Fatalf("创建会话失败: %v", err)
	}
	fmt.Printf("创建会话: %s\n", sessionID)

	// 订阅CAN ID 291
	fmt.Println("订阅CAN ID 291...")
	sub := gocan.SubscribeRequest{
		SessionID: sessionID,
		CanID:     291,
	}

	// 持续订阅
	go testSubscribe(sub)

	// 保持运行
	fmt.Println("SSE订阅已启动，按Ctrl+C退出...")
	select {} // 永久阻塞
}

func testDisconnect(sessionID string) error {
	reqBody := map[string]string{"id": sessionID}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:8080/api/can/disconnect", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result gocan.DisconnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Status != "disconnected" {
		return fmt.Errorf("断开连接失败: %s", result.Message)
	}

	return nil
}
