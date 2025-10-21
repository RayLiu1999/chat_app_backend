/**
 * 高負載測試場景 (Heavy Load Test)
 * 
 * 目的：模擬大量用戶，專注於高頻率的 API 和 WebSocket 操作
 * 特點：
 * - 大規模負載（50-200 VU）
 * - 高比例的 WebSocket 測試
 * - 高頻率的訊息發送
 * - 測試系統極限和穩定性
 * 
 * 測試重點：
 * - 30% API 操作
 * - 70% WebSocket 操作
 * - 多用戶高頻訊息發送
 * - 長時間連線穩定性
 * - 系統資源使用情況
 */
import { group, sleep, check } from 'k6';
import { randomSleep } from '../scripts/common/utils.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
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
  
  logGroupStart('Heavy Load Test - 高負載測試');
  
  // 取得認證會話
  const session = getAuthenticatedSession(baseUrl);
  
  if (!session) {
    logError('無法建立認證會話');
    logGroupEnd('Heavy Load Test (認證失敗)', testStartTime);
    return;
  }
  
  logInfo(`用戶已認證: ${session.user.email} (VU: ${__VU})`);
  
  // 30% 機率執行 API 測試，70% 機率執行 WebSocket 測試
  const randomValue = Math.random();
  
  if (randomValue < 0.3) {
    // ==================== API 測試路徑（快速執行）====================
    group('API Operations', function () {
      // 高負載下快速執行 API 操作
      const apiActions = [
        { name: 'Servers', fn: () => apiServers(baseUrl, session) },
        { name: 'Friends', fn: () => apiFriends(baseUrl, session) },
        { name: 'Chat', fn: () => apiChat(baseUrl, session) },
      ];
      
      const action = apiActions[Math.floor(Math.random() * apiActions.length)];
      logInfo(`快速執行 API 操作: ${action.name}`);
      action.fn();
      
      randomSleep(1, 2); // 最小延遲
    });
    
  } else {
    // ==================== WebSocket 測試路徑（高頻操作）====================
    group('WebSocket Operations', function () {
      logInfo('開始 WebSocket 高負載測試');
      
      if (!session.token) {
        logError('缺少 Access Token，跳過 WebSocket 測試');
        return;
      }
      
      // 使用高度共享的房間增加並發壓力
      // 每 10 個 VU 共用一個房間，限制房間數量以增加並發
      const sharedRoomIndex = Math.floor(__VU % 10);
      const testRoomId = `heavy_shared_room_${sharedRoomIndex}`;
      const roomType = 'channel';
      
      logInfo(`使用高負載共享房間 ${testRoomId} (VU: ${__VU}, Room Index: ${sharedRoomIndex})`);
      
      // 執行 WebSocket 操作
      const result = wsConnect(config.WS_URL, session.token, function (socket) {
        let messagesSent = 0;
        
        // 步驟 1: 快速加入房間
        logInfo(`步驟 1: 加入房間 ${testRoomId}`);
        joinRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(1.5); // 縮短等待時間
        
        // 步驟 2: 高頻發送訊息（5-8 條）
        const messageCount = Math.floor(Math.random() * 4) + 5; // 5-8 條
        logInfo(`步驟 2: 高頻發送 ${messageCount} 條訊息`);
        for (let i = 0; i < messageCount; i++) {
          sendMessage(socket, testRoomId, roomType);
          messagesSent++;
          sleep(0.3 + Math.random() * 0.2); // 0.3-0.5 秒快速發送
        }
        
        // 步驟 3: 持續 Ping（壓力測試）
        const pingCount = Math.floor(Math.random() * 2) + 2; // 2-3 次 ping
        logInfo(`步驟 3: 發送 ${pingCount} 次 Ping`);
        for (let i = 0; i < pingCount; i++) {
          socket.send(JSON.stringify({ action: 'ping', data: {} }));
          messagesSent++;
          sleep(0.5);
        }
        
        // 步驟 4: 繼續發送訊息
        const additionalMessages = Math.floor(Math.random() * 3) + 2; // 2-4 條額外訊息
        logInfo(`步驟 4: 發送 ${additionalMessages} 條額外訊息`);
        for (let i = 0; i < additionalMessages; i++) {
          sendMessage(socket, testRoomId, roomType);
          messagesSent++;
          sleep(0.4);
        }
        
        // 步驟 5: 離開房間
        logInfo('步驟 5: 離開房間');
        leaveRoom(socket, testRoomId, roomType);
        messagesSent++;
        sleep(1);
        
        return { messagesSent };
      }, 30);
      
      // 驗證 WebSocket 操作結果
      if (result.success) {
        logInfo(`✅ WebSocket 高負載測試完成 - 收到 ${result.receivedMessages.length} 條訊息`);
        
        // 檢查關鍵訊息
        check(result.messageStates, {
          'WS Heavy: room_joined received': (s) => s.room_joined === true,
          'WS Heavy: high frequency messages sent': (s) => (s.message_sent_count || 0) >= 5,
          'WS Heavy: pong received': (s) => (s.pong_count || 0) > 0,
        });
      } else {
        logError('❌ WebSocket 高負載測試失敗');
      }
      
      randomSleep(0.5, 1.5); // 最小延遲後立即開始下一次迭代
    });
  }
  
  const totalDuration = Date.now() - testStartTime;
  logGroupEnd('Heavy Load Test 完成', testStartTime);
  logInfo(`迭代耗時: ${(totalDuration / 1000).toFixed(2)} 秒`);
}
