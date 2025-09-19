package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go-can.com/src"
)

type MockDriver struct {
	frames []gocan.Frame
	idx    int
}

func (m *MockDriver) WriteFrame(ctx context.Context, frame gocan.Frame) error {
	fmt.Printf("Send: ID=0x%X Data=%v\n", frame.ID, frame.Data)
	return nil
}

func (m *MockDriver) ReadFrame(ctx context.Context) (gocan.Frame, error) {
	if m.idx >= len(m.frames) {
		return gocan.Frame{}, errors.New("no more frames")
	}
	frame := m.frames[m.idx]
	m.idx++
	fmt.Printf("Receive: ID=0x%X Data=%v\n", frame.ID, frame.Data)
	return frame, nil
}

func main() {
	// 加载发送帧
	sendFrames, err := gocan.LoadFramesFromFile("data/send.json")
	if err != nil {
		log.Fatalf("load send.json failed: %v", err)
	}
	// 加载接收帧
	recvFrames, err := gocan.LoadFramesFromFile("data/receive.json")
	if err != nil {
		log.Fatalf("load receive.json failed: %v", err)
	}

	driver := &MockDriver{frames: recvFrames}
	bus := gocan.NewCANBus(driver)

	// 模拟发送
	for _, frame := range sendFrames {
		if err := bus.Send(frame); err != nil {
			log.Printf("send error: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 模拟接收
	for range recvFrames {
		frame, err := bus.Receive()
		if err != nil {
			log.Printf("receive error: %v", err)
			break
		}
		_ = frame // 可扩展处理
		time.Sleep(500 * time.Millisecond)
	}
}
