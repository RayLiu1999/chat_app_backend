/**
 * 中等負載測試場景
 * 目的：模擬中等數量的用戶，混合 API 和 WebSocket 操作。
 */
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { randomSleep } from '../scripts/common/utils.js';

export default function (config) {
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  const session = getAuthenticatedSession(baseUrl);

  if (!session) return;

  // 70% 的機率執行 API 測試，30% 的機率執行 WebSocket 測試
  if (Math.random() < 0.7) {
    const apiActions = [
      () => apiServers(baseUrl, session),
      () => apiFriends(baseUrl, session),
      () => apiChat(baseUrl, session),
    ];
    const action = apiActions[Math.floor(Math.random() * apiActions.length)];
    action();
    randomSleep(2, 4);
  } else {
    const socket = wsConnect(config.WS_URL, session.token);
    if (socket) {
      const testRoomId = `room_${__VU}`;
      socket.on('open', () => {
        joinRoom(socket, testRoomId);
        for (let i = 0; i < 3; i++) {
          sendMessage(socket, testRoomId);
          randomSleep(1, 3);
        }
        leaveRoom(socket, testRoomId);
      });
      
      socket.on('close', () => {});
    }
  }
}
