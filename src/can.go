package gocan

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Frame represents a CAN frame
type Frame struct {
	ID   uint32 `json:"id"`
	Data []byte `json:"data"`
}

// LoadFramesFromFile loads frames from a JSON file
func LoadFramesFromFile(filename string) ([]Frame, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var frames []Frame
	err = json.Unmarshal(data, &frames)
	if err != nil {
		return nil, err
	}
	return frames, nil
}

// CANBus represents a CAN bus interface
type CANBus struct {
	driver Driver
}

// Driver interface for CAN communication
type Driver interface {
	WriteFrame(ctx context.Context, frame Frame) error
	ReadFrame(ctx context.Context) (Frame, error)
}

// NewCANBus creates a new CAN bus with the given driver
func NewCANBus(driver Driver) *CANBus {
	return &CANBus{driver: driver}
}

// Send sends a frame on the CAN bus
func (b *CANBus) Send(frame Frame) error {
	return b.driver.WriteFrame(context.Background(), frame)
}

// Receive receives a frame from the CAN bus
func (b *CANBus) Receive() (Frame, error) {
	return b.driver.ReadFrame(context.Background())
}

// MockDriver simulates a CAN interface
type MockDriver struct {
	frames []Frame
	idx    int
}

func (m *MockDriver) WriteFrame(ctx context.Context, frame Frame) error {
	fmt.Printf("Send: ID=0x%X Data=%v\n", frame.ID, frame.Data)
	return nil
}

func (m *MockDriver) ReadFrame(ctx context.Context) (Frame, error) {
	if m.idx >= len(m.frames) {
		return Frame{}, fmt.Errorf("no more frames")
	}
	frame := m.frames[m.idx]
	m.idx++
	fmt.Printf("Receive: ID=0x%X Data=%v\n", frame.ID, frame.Data)
	return frame, nil
}
