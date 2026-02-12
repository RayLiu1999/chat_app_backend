/**
 * 單體容量測試場景 (Monolith Capacity Test)
 * 
 * 模擬真實業務比例 (Soak Test 模式)：
 * - 90% WebSocket 長連線 (持久連線、心跳、低頻發言)
 * - 10% API 操作 (讀取伺服器、好友、訊息)
 * - 移除重複登入 (僅在無 Session 時登入)
 */
import { group, sleep } from 'k6';
import { randomSleep } from '../scripts/common/utils.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { logInfo, logError, logGroupStart, logGroupEnd } from '../scripts/common/logger.js';

// VU 級別的狀態快取
let vuSession = null;
let vuRoomId = null;

export default function (config, preAuthSession) {
  const testStartTime = Date.now();
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  
  logGroupStart('Monolith Capacity Test - 單體容量測試');
  
  // 1. 確保有 Session (每個 VU 使用獨立的 Session)
  if (!vuSession) {
    const jitter = Math.random() * 5000;
    sleep(jitter / 1000);
    logInfo(`[Auth] VU ${__VU} 正在初始化獨立 Session (Jitter: ${jitter.toFixed(0)}ms)`);
    vuSession = getAuthenticatedSession(baseUrl);
  }
  
  const session = vuSession;
  if (!session || !session.token) {
    logError(`[Auth] VU ${__VU} 無法取得有效 Session，跳過測試`);
    return;
  }

  // 2. 確保有測試房間 (每個動態用戶需要一個家)
  if (!vuRoomId) {
    logInfo(`[Env] VU ${__VU} 正在獲取測試房間...`);
    try {
      // 呼叫伺服器 API 腳本，它會自動：獲取現有 -> 若無則創建
      const serverData = apiServers(baseUrl, session);
      if (serverData && serverData.channelId) {
        vuRoomId = serverData.channelId;
        logInfo(`[Env] VU ${__VU} 已就緒，使用房間: ${vuRoomId}`);
      } else {
        logError(`[Env] VU ${__VU} 無法取得測試房間，跳過本次迭代`);
        vuRoomId = null;
      }
    } catch (e) {
      logError(`[Env] 獲取房間失敗: ${e.message}`);
      vuRoomId = null;
    }
  }

  if (!vuRoomId) {
    logError(`[WS] VU ${__VU} 無可用房間，跳過本次迭代`);
    sleep(5);
    return;
  }

  const randomValue = Math.random();
  
  // 3. 10% API 讀取操作
  if (randomValue < 0.1) {
    group('API Read Operations', function () {
      const actions = [
        { name: 'Get Servers', fn: () => apiServers(baseUrl, session) },
        { name: 'Get Friends', fn: () => apiFriends(baseUrl, session) },
        { name: 'Get Chat History', fn: () => apiChat(baseUrl, session) },
      ];
      
      const actionIdx = Math.floor(Math.random() * actions.length);
      const action = actions[actionIdx];
      logInfo(`[API] 執行操作: ${action.name} (VU: ${__VU})`);
      action.fn();
      randomSleep(1, 2);
    });
  }
  // 4. 90% WebSocket 長連線 (模擬掛網聊天)
  else {
    group('WebSocket Persistent Connection', function () {
      const testRoomId = vuRoomId;
      const roomType = 'channel';

      logInfo(`[WS] 連線到房間: ${testRoomId} (VU: ${__VU})`);

      // 建立連線，使用自定義 Handler 執行長連線邏輯
      wsConnect(config.WS_URL, session.token, function (socket) {
        // 1. 加入房間 (只做一次)
        joinRoom(socket, testRoomId, roomType);
        
        // 2. 長時間掛網循環
        const duration = 5 * 60 * 1000; 
        const endTime = Date.now() + duration;
        
        logInfo(`[WS] 開始長連線掛網模式 (${duration/1000}s)`);

        while (Date.now() < endTime) {
          socket.send(JSON.stringify({ action: 'ping', data: {} }));

          if (Math.random() < 0.3) { 
             sendMessage(socket, testRoomId, roomType);
          }
          
          sleep(10 + Math.random() * 20); // 休息 10~30 秒
        }

        // 3. 測試結束時離開
        leaveRoom(socket, testRoomId, roomType);
        sleep(0.5);
      }, 360); 
    });
  }
  
  logGroupEnd('Monolith Capacity Test', testStartTime);
}
