/**
 * 單體混合壓力測試場景 (Monolith Mixed Stress Test)
 * 
 * 模擬真實業務比例：
 * - 20% 登入 (Bcrypt 壓力)
 * - 30% API 操作 (讀取伺服器、好友、訊息)
 * - 50% WebSocket 即時通訊 (持久連線、訊息發送、心跳)
 */
import { group, sleep, check } from 'k6';
import { randomSleep } from '../scripts/common/utils.js';
import { login, getAuthenticatedSession } from '../scripts/common/auth.js';
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { logInfo, logError, logGroupStart, logGroupEnd } from '../scripts/common/logger.js';

export default function (config, preAuthSession) {
  const testStartTime = Date.now();
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  
  logGroupStart('Monolith Mixed Test - 單體混合測試');
  
  const randomValue = Math.random();
  
  // 1. 20% 登入測試 (Bcrypt 關鍵瓶頸驗證)
  if (randomValue < 0.2) {
    group('Auth Operations (Bcrypt Focus)', function () {
      logInfo(`[Auth] 執行登入測試 (VU: ${__VU})`);
      const user = config.TEST_USERS[__VU % config.TEST_USERS.length];
      const result = login(baseUrl, {
        email: user.email,
        password: user.password,
      });
      
      check(result, {
        'Login success': (r) => r && r.token !== undefined,
      });
      randomSleep(1, 3);
    });
  } 
  // 2. 30% API 讀取操作
  else if (randomValue < 0.5) {
    group('API Read Operations', function () {
      const session = preAuthSession;
      if (!session) return;

      const actions = [
        { name: 'Get Servers', fn: () => apiServers(baseUrl, session) },
        { name: 'Get Friends', fn: () => apiFriends(baseUrl, session) },
        { name: 'Get Chat History', fn: () => apiChat(baseUrl, session) },
      ];
      
      const action = actions[Math.floor(Math.random() * actions.length)];
      logInfo(`[API] 執行操作: ${action.name} (VU: ${__VU})`);
      action.fn();
      randomSleep(1, 2);
    });
  }
  // 3. 50% WebSocket 即時通訊
  else {
    group('WebSocket Operations', function () {
      const session = preAuthSession;
      if (!session || !session.token) {
        logError('缺少 Session，跳過 WS 測試');
        return;
      }

      // 每 5 個 VU 共用一個測試房間
      // 使用符合 ObjectID 格式的種子 ID (24位16進制)
      const roomIndex = Math.floor(__VU % 5);
      const testRoomId = `00000000000000000000000${roomIndex}`;
      const roomType = 'channel';

      logInfo(`[WS] 連線到房間: ${testRoomId} (VU: ${__VU})`);

      const result = wsConnect(config.WS_URL, session.token, function (socket) {
        // 加入房間
        joinRoom(socket, testRoomId, roomType);
        sleep(1);

        // 發送 2-4 條訊息
        const msgCount = Math.floor(Math.random() * 3) + 2;
        for (let i = 0; i < msgCount; i++) {
          sendMessage(socket, testRoomId, roomType);
          sleep(0.5 + Math.random());
        }

        // 心跳測試
        socket.send(JSON.stringify({ action: 'ping', data: {} }));
        sleep(1);

        // 離開房間
        leaveRoom(socket, testRoomId, roomType);
        sleep(0.5);
      }, 20); // 持續 20 秒

      check(result, {
        'WS success': (r) => r.success === true,
      });
    });
  }
  
  logGroupEnd('Monolith Mixed Test', testStartTime);
}
