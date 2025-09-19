#!/bin/bash

# CAN API 测试脚本
echo "=== CAN API 测试脚本 ==="

# 测试连接API
echo -e "\n1. 测试连接API..."
response=$(curl -s -X POST http://localhost:8080/api/can/connect \
  -H "Content-Type: application/json" \
  -d '{"channel":"PCAN_USBBUS1"}')

echo "响应: $response"

# 提取session ID
session_id=$(echo $response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Session ID: $session_id"

# 测试发送API
echo -e "\n2. 测试发送API..."
send_response=$(curl -s -X POST http://localhost:8080/api/can/send \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$session_id\",\"frame\":{\"id\":291,\"data\":[1,2,3,4]}}")

echo "发送响应: $send_response"

# 测试断开连接API
echo -e "\n3. 测试断开连接API..."
disconnect_response=$(curl -s -X POST http://localhost:8080/api/can/disconnect \
  -H "Content-Type: application/json" \
  -d "{\"id\":\"$session_id\"}")

echo "断开连接响应: $disconnect_response"

echo -e "\n=== 测试完成 ==="
