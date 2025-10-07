/**
 * 高負載測試場景
 * 目的：模擬大量用戶，專注於高頻率的 API 和 WebSocket 操作。
 */
import apiServers from '../scripts/api/servers.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { randomSleep } from '../scripts/common/utils.js';

export default function (config) {
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  const session = getAuthenticatedSession(baseUrl);

  if (!session) return;

  // 50% API, 50% WebSocket
  if (Math.random() < 0.5) {
    apiServers(baseUrl, session);
    randomSleep(1, 2);
  } else {
    const socket = wsConnect(config.WS_URL, session.token);
    if (socket) {
      const testRoomId = `heavy_room_${__VU % 10}`; // 限制房間數量以增加並發
      socket.on('open', () => {
        joinRoom(socket, testRoomId);
        for (let i = 0; i < 5; i++) {
          sendMessage(socket, testRoomId);
          randomSleep(0.5, 1.5);
        }
        leaveRoom(socket, testRoomId);
      });
      socket.on('close', () => {});
    }
  }
}
