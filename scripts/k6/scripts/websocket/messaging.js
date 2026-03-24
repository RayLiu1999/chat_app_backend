/**
 * WebSocket 測試 - 訊息發送 (Messaging)
 * 注意：後端使用 'action' 而不是 'type'
 */
import { check } from 'k6';
import { randomItem } from '../common/utils.js';
import { logInfo } from '../common/logger.js';
import { SharedArray } from 'k6/data';

const messages = new SharedArray('messages', function () {
  try {
    // 相對於本檔案位置: loadtest/scripts/websocket/messaging.js → loadtest/data/messages.json
    return JSON.parse(open('../../data/messages.json'));
  } catch (e) {
    return ['Hello', 'Test message', 'k6 ws load test', 'WebSocket test'];
  }
});

// 根據後端實際的 WebSocket API 格式發送訊息
export function sendMessage(socket, roomId, roomType = 'dm') {
  if (!socket) return;
  const messageContent = randomItem(messages);
  
  logInfo(`📨 發送訊息到房間: ${roomType}:${roomId}`, { content: messageContent });
  
  const payload = JSON.stringify({
    action: 'send_message',  // 使用 action 而不是 type
    data: {
      room_id: roomId,
      room_type: roomType, // 'dm' 或 'channel'
      content: messageContent,
    },
  });
  socket.send(payload);
}

// 加入房間
export function joinRoom(socket, roomId, roomType = 'dm') {
  if (!socket) return;
  
  logInfo(`📨 發送加入房間請求: ${roomType}:${roomId}`);
  
  const payload = JSON.stringify({
    action: 'join_room',  // 使用 action 而不是 type
    data: {
      room_id: roomId,
      room_type: roomType,
    },
  });
  socket.send(payload);
}

// 離開房間
export function leaveRoom(socket, roomId, roomType = 'dm') {
  if (!socket) return;
  
  logInfo(`📨 發送離開房間請求: ${roomType}:${roomId}`);
  
  const payload = JSON.stringify({
    action: 'leave_room',  // 使用 action 而不是 type
    data: {
      room_id: roomId,
      room_type: roomType,
    },
  });
  socket.send(payload);
}
