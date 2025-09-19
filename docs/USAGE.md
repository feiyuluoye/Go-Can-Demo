# CAN Bus Simulation Usage Guide

## Overview
This program simulates CAN (Controller Area Network) bus communication by sending and receiving frames defined in JSON configuration files.

## Data Format Explanation
- **ID**: 32-bit message identifier (hexadecimal format). Determines message priority and type.
  - Example: `0x123` (291 decimal) might represent engine data in automotive systems
  - Higher priority messages use lower ID values
- **Data**: Byte array payload (up to 8 bytes in standard CAN 2.0 frames)
  - Example: `[1, 2, 3, 4]` could represent:
    - RPM = 258 (1*256 + 2)
    - Temperature = 768 (3*256 + 4)

## Configuration Files
### send.json
Contains frames to be transmitted:
```json
[
  {
    "id": 291,
    "data": [1, 2, 3, 4]
  },
  {
    "id": 292,
    "data": [5, 6, 7, 8]
  }
]
```

### receive.json
Contains frames to be received by the mock driver:
```json
[
  {
    "id": 291,
    "data": [9, 10, 11, 12]
  },
  {
    "id": 292,
    "data": [13, 14, 15, 16]
  }
]
```

## Running the Program
1. Ensure Go 1.25+ is installed
2. Execute: `go run main.go`
3. Expected output shows sent/received frames:
```
Send: ID=0x123 Data=[1 2 3 4]
Send: ID=0x124 Data=[5 6 7 8]
Receive: ID=0x123 Data=[9 10 11 12]
Receive: ID=0x124 Data=[13 14 15 16]
```

## Customization
- Modify `src/send.json`/`src/receive.json` to simulate different scenarios
- Extend `src/can.go` to add real hardware drivers
- Implement additional CAN protocols (J1939, CANopen) by extending the Frame structure