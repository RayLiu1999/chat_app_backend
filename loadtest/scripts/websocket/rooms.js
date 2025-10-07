/**
 * WebSocket 測試 - 房間管理 (Rooms)
 * 注意：後端使用 'action' 而不是 'type'
 */
import { check } from 'k6';
import { logInfo } from '../common/logger.js';

export function joinRoom(socket, roomId, roomType = 'channel') {
  if (!socket) return;
  
  logInfo(`📨 發送加入房間請求: ${roomType}:${roomId}`);
  
  const payload = JSON.stringify({
    action: 'join_room',  // 使用 action 而不是 type
    data: { room_id: roomId, room_type: roomType },
  });
  socket.send(payload);
}

export function leaveRoom(socket, roomId, roomType = 'channel') {
  if (!socket) return;
  
  logInfo(`📨 發送離開房間請求: ${roomType}:${roomId}`);
  
  const payload = JSON.stringify({
    action: 'leave_room',  // 使用 action 而不是 type
    data: { room_id: roomId, room_type: roomType },
  });
  socket.send(payload);
}
