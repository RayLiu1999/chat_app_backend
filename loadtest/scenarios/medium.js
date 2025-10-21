/**
 * 中等負載測試場景 (Medium Load Test)
 * 
 * 目的：模擬中等數量的用戶，混合 API 和 WebSocket 操作
 * 特點：
 * - 中等規模負載（20-50 VU）
 * - 平衡的 API 和 WebSocket 測試比例
 * - 較高的操作頻率
 * - 測試多房間並發場景
 * 
 * 測試重點：
 * - 50% API 操作
 * - 50% WebSocket 操作
 * - 多用戶同時在同一房間
 * - 較高頻率的訊息發送
 */
import { group, sleep, check } from 'k6';
import { randomSleep } from '../scripts/common/utils.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import apiAuth from '../scripts/api/auth.js';
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import wsConnect from '../scripts/websocket/connection.js';
import { joinRoom, leaveRoom } from '../scripts/websocket/rooms.js';
import { sendMessage } from '../scripts/websocket/messaging.js';
import { logInfo, logError, logGroupStart, logGroupEnd } from '../scripts/common/logger.js';

export default function (config) {
  const testStartTime = Date.now();
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;
  
  logGroupStart('Medium Load Test - 中等負載測試');
  
  // 取得認證會話
  const session = getAuthenticatedSession(baseUrl);
  
  if (!session) {
    logError('無法建立認證會話');
    logGroupEnd('Medium Load Test (認證失敗)', testStartTime);
    return;
  }
  
  logInfo(`用戶已認證: ${session.user.email} (VU: ${__VU})`);
  
  // 50% 機率執行 API 測試，50% 機率執行 WebSocket 測試
  const randomValue = Math.random();
  
  if (randomValue < 0.5) {
    // ==================== API 測試路徑 ====================
    group('API Operations', function () {
      // 中等負載下執行 1-2 個 API 操作
      const operationCount = Math.floor(Math.random() * 2) + 1; // 1-2 個操作
      
      const apiActions = [
        { name: 'Servers', fn: () => apiServers(baseUrl, session) },
        { name: 'Friends', fn: () => apiFriends(baseUrl, session) },
        { name: 'Chat', fn: () => apiChat(baseUrl, session) },
      ];
      
      for (let i = 0; i < operationCount; i++) {
        const action = apiActions[Math.floor(Math.random() * apiActions.length)];
        logInfo(`執行 API 操作 ${i + 1}/${operationCount}: ${action.name}`);
        action.fn();
        randomSleep(1, 2); // 操作之間的短延遲
      }
      
      randomSleep(2, 4); // 結束前的延遲
    });
    
  } else {
    // ==================== WebSocket 測試路徑 ====================
    group('WebSocket Operations', function () {
      logInfo('開始 WebSocket 測試');
      
      if (!session.token) {
        logError('缺少 Access Token，跳過 WebSocket 測試');
        return;
      }
      
      // 使用共享房間增加並發壓力
      // 每 5 個 VU 共用一個房間
      const sharedRoomIndex = Math.floor(__VU / 5);
      const testRoomId = `medium_shared_room_${sharedRoomIndex}`;
      const roomType = 'channel';
      
      logInfo(`使用共享房間 ${testRoomId} (VU: ${__VU}, Room Index: ${sharedRoomIndex})`);
      
      // 執行 WebSocket 操作
      const result = wsConnect(config.WS_URL, session.token, function (socket) {
        let messagesSent = 0;
        
        // 步驟 1: 加入共享房間
        logInfo(`步驟 1: 加入共享房間 ${testRoomId}`);
        joinRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(2);
        
        // 步驟 2: 發送 3-5 條訊息（中等負載）
        const messageCount = Math.floor(Math.random() * 3) + 3; // 3-5 條
        logInfo(`步驟 2: 發送 ${messageCount} 條訊息`);
        for (let i = 0; i < messageCount; i++) {
          sendMessage(socket, testRoomId, roomType);
          messagesSent++;
          sleep(0.5 + Math.random() * 0.5); // 0.5-1 秒隨機延遲
        }
        
        // 步驟 3: 隨機執行 Ping 測試
        if (Math.random() < 0.7) {
          logInfo('步驟 3: 測試 Ping/Pong');
          socket.send(JSON.stringify({ action: 'ping', data: {} }));
          messagesSent++;
          sleep(1);
        }
        
        // 步驟 4: 保持連線一段時間（模擬真實用戶）
        const stayDuration = Math.floor(Math.random() * 3) + 2; // 2-4 秒
        logInfo(`步驟 4: 保持連線 ${stayDuration} 秒`);
        sleep(stayDuration);
        
        // 步驟 5: 離開房間
        logInfo('步驟 5: 離開房間');
        leaveRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(1);
        
        return { messagesSent };
      }, 25);
      
      // 驗證 WebSocket 操作結果
      if (result.success) {
        logInfo(`✅ WebSocket 測試完成 - 收到 ${result.receivedMessages.length} 條訊息`);
        
        // 檢查關鍵訊息
        check(result.messageStates, {
          'WS Medium: room_joined received': (s) => s.room_joined === true,
          'WS Medium: message operations completed': (s) => (s.message_sent_count || 0) > 0,
        });
      } else {
        logError('❌ WebSocket 測試失敗');
      }
      
      randomSleep(1, 3); // 結束前的延遲
    });
  }
  
  const totalDuration = Date.now() - testStartTime;
  logGroupEnd('Medium Load Test 完成', testStartTime);
  logInfo(`迭代耗時: ${(totalDuration / 1000).toFixed(2)} 秒`);
}
