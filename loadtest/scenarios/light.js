/**
 * 輕量負載測試場景 (Light Load Test)
 * 
 * 目的：模擬少量用戶同時在線，進行常規操作
 * 特點：
 * - 小規模負載（5-10 VU）
 * - 混合 API 和 WebSocket 測試
 * - 適度的操作頻率
 * - 涵蓋所有核心功能
 * 
 * 測試重點：
 * - 70% API 操作 (認證、伺服器、好友、聊天)
 * - 30% WebSocket 操作 (連線、房間、訊息)
 */
import { group, sleep, check } from 'k6';
import { randomSleep } from '../scripts/common/utils.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import apiAuth from '../scripts/api/auth.js';
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import apiUpload from '../scripts/api/upload.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { logInfo, logError, logGroupStart, logGroupEnd } from '../scripts/common/logger.js';

export default function (config) {
  const testStartTime = Date.now();
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  
  logGroupStart('Light Load Test - 輕量負載測試');
  
  // 取得認證會話
  const session = getAuthenticatedSession(baseUrl);
  
  if (!session) {
    logError('無法建立認證會話，嘗試執行認證測試');
    apiAuth(baseUrl);
    logGroupEnd('Light Load Test (認證失敗)', testStartTime);
    return;
  }
  
  logInfo(`用戶已認證: ${session.user.email}`);
  
  // 70% 機率執行 API 測試，30% 機率執行 WebSocket 測試
  const randomValue = Math.random();
  
  if (randomValue < 0.7) {
    // ==================== API 測試路徑 ====================
    group('API Operations', function () {
      // 隨機選擇一個 API 操作
      const apiActions = [
        { name: 'Servers', fn: () => apiServers(baseUrl, session), weight: 0.3 },
        { name: 'Friends', fn: () => apiFriends(baseUrl, session), weight: 0.3 },
        { name: 'Chat', fn: () => apiChat(baseUrl, session), weight: 0.3 },
        { name: 'Upload', fn: () => apiUpload(baseUrl, session), weight: 0.1 },
      ];
      
      // 加權隨機選擇
      const totalWeight = apiActions.reduce((sum, action) => sum + action.weight, 0);
      let random = Math.random() * totalWeight;
      let selectedAction = apiActions[0];
      
      for (const action of apiActions) {
        random -= action.weight;
        if (random <= 0) {
          selectedAction = action;
          break;
        }
      }
      
      logInfo(`執行 API 操作: ${selectedAction.name}`);
      selectedAction.fn();
      randomSleep(2, 5); // API 操作之間的延遲
    });
    
  } else {
    // ==================== WebSocket 測試路徑 ====================
    group('WebSocket Operations', function () {
      logInfo('開始 WebSocket 測試');
      
      if (!session.token) {
        logError('缺少 Access Token，跳過 WebSocket 測試');
        return;
      }
      
      // 先取得房間資訊
      let testRoomId;
      let roomType = 'channel';
      
      // 嘗試取得伺服器和頻道資訊
      const serverData = apiServers(baseUrl, session);
      if (serverData && serverData.channelId) {
        testRoomId = serverData.channelId;
        roomType = 'channel';
        logInfo(`使用頻道進行 WebSocket 測試: ${testRoomId}`);
      } else {
        // 如果沒有頻道，使用虛擬房間 ID
        testRoomId = `light_test_room_${__VU}`;
        logInfo(`使用虛擬房間進行 WebSocket 測試: ${testRoomId}`);
      }
      
      // 執行 WebSocket 操作
      const result = wsConnect(config.WS_URL, session.token, function (socket) {
        let messagesSent = 0;
        
        // 步驟 1: 加入房間
        logInfo(`步驟 1: 加入房間 ${testRoomId}`);
        joinRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(2);
        
        // 步驟 2: 發送 2-3 條訊息（輕量測試）
        const messageCount = Math.floor(Math.random() * 2) + 2; // 2-3 條
        logInfo(`步驟 2: 發送 ${messageCount} 條訊息`);
        for (let i = 0; i < messageCount; i++) {
          sendMessage(socket, testRoomId, roomType);
          messagesSent++;
          sleep(1);
        }
        
        // 步驟 3: Ping 測試（可選）
        if (Math.random() < 0.5) {
          logInfo('步驟 3: 測試 Ping/Pong');
          socket.send(JSON.stringify({ action: 'ping', data: {} }));
          messagesSent++;
          sleep(1);
        }
        
        // 步驟 4: 離開房間
        logInfo('步驟 4: 離開房間');
        leaveRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(1);
        
        return { messagesSent };
      }, 20);
      
      // 驗證 WebSocket 操作結果
      if (result.success) {
        logInfo(`✅ WebSocket 測試完成 - 收到 ${result.receivedMessages.length} 條訊息`);
      } else {
        logError('❌ WebSocket 測試失敗');
      }
      
      randomSleep(2, 4); // WebSocket 操作之間的延遲
    });
  }
  
  const totalDuration = Date.now() - testStartTime;
  logGroupEnd('Light Load Test 完成', testStartTime);
  logInfo(`迭代耗時: ${(totalDuration / 1000).toFixed(2)} 秒`);
}
