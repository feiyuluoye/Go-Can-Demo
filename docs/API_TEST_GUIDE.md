# CAN API 测试指南

基于README.md第四章的API规范，使用Gin框架实现的CAN总线通信API接口。

## API端点

### 1. 建立CAN连接
**POST** `/api/can/connect`

请求体：
```json
{
  "channel": "PCAN_USBBUS1"
}
```

响应：
```json
{
  "status": "connected",
  "id": "session_1234567890",
  "message": "Connected to channel PCAN_USBBUS1"
}
```

### 2. 断开CAN连接
**POST** `/api/can/disconnect`

请求体：
```json
{
  "id": "session_1234567890"
}
```

响应：
```json
{
  "status": "disconnected",
  "message": "Session session_1234567890 disconnected"
}
```

### 3. 发送CAN帧
**POST** `/api/can/send`

请求体：
```json
{
  "id": "session_1234567890",
  "frame": {
    "id": 291,
    "data": [1, 2, 3, 4]
  }
}
```

响应：
```json
{
  "status": "sent",
  "timestamp": "2025-09-19T15:18:00.0806157+08:00",
  "message": "Frame sent: ID=0x123"
}
```

### 4. 订阅CAN消息 (SSE流)
**GET** `/api/can/subscribe`

请求体：
```json
{
  "id": "session_1234567890",
  "canId": 291
}
```

响应：Server-Sent Events流式数据

## Mock数据文件

已自动生成以下mock数据文件：

### data/connect_mock.json
```json
[
  {"channel": "PCAN_USBBUS1"},
  {"channel": "PCAN_USBBUS2"},
  {"channel": "can0"}
]
```

### data/send_mock.json
```json
[
  {
    "id": "session_123",
    "frame": {
      "id": 291,
      "data": "AQIDBA=="  // Base64编码的 [1,2,3,4]
    }
  }
]
```

### data/subscribe_mock.json
```json
[
  {"id": "session_123", "canId": 291},
  {"id": "session_123", "canId": 292}
]
```

## 测试方法

### 1. 启动服务器
```bash
go run main.go
```

### 2. 使用curl测试
```bash
# 建立连接
curl http://localhost:8080/api/can/connect \
  -H "Content-Type: application/json" \
  -d '{"channel":"PCAN_USBBUS1"}'

# 发送CAN帧
curl http://localhost:8080/api/can/send \
  -H "Content-Type: application/json" \
  -d '{"id":"session_123","frame":{"id":291,"data":[1,2,3,4]}}'

# 断开连接
curl http://localhost:8080/api/can/disconnect \
  -H "Content-Type: application/json" \
  -d '{"id":"session_123"}'
```

### 3. 使用测试脚本
```bash
chmod +x test_api.sh
./test_api.sh
```

### 4. 测试SSE订阅
使用浏览器或专门的SSE客户端测试：
```javascript
// 浏览器中测试
const eventSource = new EventSource('http://localhost:8080/api/can/subscribe?id=session_123&canId=291');
eventSource.onmessage = function(event) {
  console.log('Received:', event.data);
};
```

## 技术实现

- **框架**: Gin Web Framework
- **数据格式**: JSON
- **流式传输**: Server-Sent Events (SSE)
- **会话管理**: 内存中会话存储
- **Mock驱动**: 模拟CAN总线通信

## 注意事项

1. 当前为模拟环境，使用MockDriver模拟CAN通信
2. Session ID由服务器自动生成
3. SSE连接需要保持长连接
4. 数据字段中的byte数组会自动进行Base64编码
