/**
 * WebSocket 壓力測試場景
 * 
 * 目的：專門測試 WebSocket 連線的承載能力和高頻訊息處理
 * 
 * 測試重點：
 * 1. 大量並發 WebSocket 連線
 * 2. 高頻率訊息發送
 * 3. 多房間並發操作
 * 4. 連線穩定性測試
 * 
 * 使用方法：
 * k6 run run.js --env SCENARIO=websocket_stress --env WS_TEST_TYPE=connections
 * k6 run run.js --env SCENARIO=websocket_stress --env WS_TEST_TYPE=messaging
 * k6 run run.js --env SCENARIO=websocket_stress --env WS_TEST_TYPE=mixed
 */

import { check, sleep, group } from 'k6';
import ws from 'k6/ws';
import { Counter, Trend, Rate } from 'k6/metrics';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { logInfo, logError, logSuccess, logGroupStart, logGroupEnd } from '../scripts/common/logger.js';

// 自定義 WebSocket 指標
const wsConnectionDuration = new Trend('ws_connection_duration');
const wsMessageSendDuration = new Trend('ws_message_send_duration');
const wsMessageReceiveRate = new Rate('ws_message_receive_rate');
const wsConnectionErrors = new Counter('ws_connection_errors');
const wsActiveConnections = new Counter('ws_active_connections');

/**
 * WebSocket 連線壓力測試
 * 目標：測試系統能承載多少並發 WebSocket 連線
 */
function testConnectionStress(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const connectionStart = Date.now();
  
  logInfo(`🔌 VU ${__VU} 開始建立 WebSocket 連線`);

  const response = ws.connect(wsUrl, {
    headers: {
      'Authorization': `Bearer ${session.token}`,
    },
    tags: { test_type: 'connection_stress' },
  }, function (socket) {
    const connectionTime = Date.now() - connectionStart;
    wsConnectionDuration.add(connectionTime);
    wsActiveConnections.add(1);
    
    logSuccess(`WebSocket 連線成功`, 101, connectionTime);

    socket.on('open', () => {
      logInfo('✅ WebSocket 已開啟');
      
      // 加入一個共用房間（模擬多人同時在線的場景）
      const sharedRoomId = `stress_room_${Math.floor(__VU / 10)}`; // 每 10 個 VU 共用一個房間
      
      socket.send(JSON.stringify({
        action: 'join_room',
        room_id: sharedRoomId
      }));
      
      logInfo(`📥 加入壓力測試房間: ${sharedRoomId}`);
    });

    socket.on('message', (data) => {
      try {
        const message = JSON.parse(data);
        logInfo(`📨 收到訊息: ${message.action || message.type}`);
        
        if (message.action === 'error' || message.type === 'error') {
          logError(`收到錯誤訊息: ${message.message}`);
        }
        
        wsMessageReceiveRate.add(1);
      } catch (e) {
        logError(`解析訊息失敗: ${e.message}`);
      }
    });

    socket.on('error', (e) => {
      logError(`WebSocket 錯誤: ${e.error()}`);
      wsConnectionErrors.add(1);
    });

    socket.on('close', () => {
      logInfo('🔴 WebSocket 連線已關閉');
    });

    // 保持連線活躍，模擬真實用戶在線
    const keepAliveInterval = 10; // 每 10 秒發送一次 ping
    for (let i = 0; i < 30; i++) { // 保持連線 5 分鐘
      socket.setTimeout(() => {
        if (socket.readyState === 1) { // OPEN
          socket.send(JSON.stringify({ action: 'ping' }));
        }
      }, i * keepAliveInterval * 1000);
    }

    // 在測試結束前保持連線
    socket.setTimeout(() => {
      socket.close();
    }, 300000); // 5 分鐘後關閉
  });

  check(response, {
    'WebSocket 連線成功': (r) => r && r.status === 101,
  });
}

/**
 * WebSocket 高頻訊息測試
 * 目標：測試系統處理高頻率訊息的能力
 */
function testMessagingStress(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  
  logInfo(`💬 VU ${__VU} 開始高頻訊息測試`);

  const response = ws.connect(wsUrl, {
    headers: {
      'Authorization': `Bearer ${session.token}`,
    },
    tags: { test_type: 'messaging_stress' },
  }, function (socket) {
    logSuccess('WebSocket 連線成功 - 開始高頻訊息測試');

    let messageCount = 0;
    let roomJoined = false;
    const roomId = `high_freq_room_${__VU % 5}`; // 5 個高頻房間

    socket.on('open', () => {
      // 先加入房間
      socket.send(JSON.stringify({
        action: 'join_room',
        room_id: roomId
      }));
    });

    socket.on('message', (data) => {
      try {
        const message = JSON.parse(data);
        
        // 檢查是否成功加入房間
        if (message.action === 'status' && message.message && message.message.includes('加入房間成功')) {
          roomJoined = true;
          logInfo(`✅ 成功加入高頻測試房間: ${roomId}`);
          
          // 開始發送高頻訊息
          startHighFrequencyMessaging(socket, roomId);
        }
        
        if (message.action === 'new_message') {
          wsMessageReceiveRate.add(1);
        }
      } catch (e) {
        logError(`解析訊息失敗: ${e.message}`);
      }
    });

    socket.on('error', (e) => {
      logError(`WebSocket 錯誤: ${e.error()}`);
      wsConnectionErrors.add(1);
    });

    socket.on('close', () => {
      logInfo(`🔴 高頻訊息測試完成，共發送 ${messageCount} 條訊息`);
    });

    // 高頻發送訊息
    function startHighFrequencyMessaging(socket, roomId) {
      const messagesPerSecond = 10; // 每秒 10 條訊息
      const testDuration = 30; // 測試 30 秒
      const totalMessages = messagesPerSecond * testDuration;
      const interval = 1000 / messagesPerSecond; // 每條訊息間隔

      for (let i = 0; i < totalMessages; i++) {
        socket.setTimeout(() => {
          if (socket.readyState === 1) {
            const sendStart = Date.now();
            
            socket.send(JSON.stringify({
              action: 'send_message',
              room_id: roomId,
              content: `高頻測試訊息 #${i + 1} from VU ${__VU}`,
              message_type: 'text'
            }));
            
            const sendDuration = Date.now() - sendStart;
            wsMessageSendDuration.add(sendDuration);
            messageCount++;
            
            if ((i + 1) % 50 === 0) {
              logInfo(`📤 已發送 ${i + 1}/${totalMessages} 條訊息`);
            }
          }
        }, i * interval);
      }

      // 測試結束後關閉連線
      socket.setTimeout(() => {
        socket.close();
      }, (testDuration + 5) * 1000);
    }
  });

  check(response, {
    'WebSocket 連線成功': (r) => r && r.status === 101,
  });
}

/**
 * 混合壓力測試
 * 目標：模擬真實場景，包含連線、多房間切換、訊息發送等
 */
function testMixedStress(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  
  logInfo(`🔀 VU ${__VU} 開始混合壓力測試`);

  const response = ws.connect(wsUrl, {
    headers: {
      'Authorization': `Bearer ${session.token}`,
    },
    tags: { test_type: 'mixed_stress' },
  }, function (socket) {
    logSuccess('WebSocket 連線成功 - 開始混合測試');

    const rooms = [
      `mixed_room_1`,
      `mixed_room_2`,
      `mixed_room_3`,
    ];
    
    let currentRoomIndex = 0;
    let messagesSent = 0;

    socket.on('open', () => {
      // 開始房間跳轉和訊息發送循環
      startMixedOperations(socket);
    });

    socket.on('message', (data) => {
      try {
        const message = JSON.parse(data);
        
        if (message.action === 'new_message') {
          wsMessageReceiveRate.add(1);
        }
      } catch (e) {
        logError(`解析訊息失敗: ${e.message}`);
      }
    });

    socket.on('error', (e) => {
      logError(`WebSocket 錯誤: ${e.error()}`);
      wsConnectionErrors.add(1);
    });

    socket.on('close', () => {
      logInfo(`🔴 混合測試完成，切換了 ${currentRoomIndex + 1} 個房間，發送了 ${messagesSent} 條訊息`);
    });

    function startMixedOperations(socket) {
      let operationIndex = 0;
      const totalOperations = 60; // 總共 60 次操作
      
      const performOperation = () => {
        if (operationIndex >= totalOperations || socket.readyState !== 1) {
          socket.close();
          return;
        }

        const operation = operationIndex % 3;
        
        switch (operation) {
          case 0: // 加入房間
            const roomToJoin = rooms[currentRoomIndex % rooms.length];
            socket.send(JSON.stringify({
              action: 'join_room',
              room_id: roomToJoin
            }));
            logInfo(`📥 加入房間: ${roomToJoin}`);
            break;
            
          case 1: // 發送訊息
            const currentRoom = rooms[currentRoomIndex % rooms.length];
            socket.send(JSON.stringify({
              action: 'send_message',
              room_id: currentRoom,
              content: `混合測試訊息 #${messagesSent + 1} from VU ${__VU}`,
              message_type: 'text'
            }));
            messagesSent++;
            break;
            
          case 2: // 離開房間並切換
            const roomToLeave = rooms[currentRoomIndex % rooms.length];
            socket.send(JSON.stringify({
              action: 'leave_room',
              room_id: roomToLeave
            }));
            logInfo(`📤 離開房間: ${roomToLeave}`);
            currentRoomIndex++;
            break;
        }
        
        operationIndex++;
        
        // 每秒 2-3 次操作
        const nextDelay = 300 + Math.random() * 200;
        socket.setTimeout(performOperation, nextDelay);
      };
      
      // 開始第一次操作
      performOperation();
    }
  });

  check(response, {
    'WebSocket 連線成功': (r) => r && r.status === 101,
  });
}

/**
 * 主測試函數
 */
export default function (config) {
  const testType = __ENV.WS_TEST_TYPE || 'mixed'; // connections, messaging, mixed
  
  logGroupStart(`WebSocket 壓力測試 - 類型: ${testType}`);
  
  // 取得認證會話
  const session = getAuthenticatedSession(`${config.BASE_URL}${config.API_PREFIX}`);
  
  if (!session) {
    logError('⚠️  無法建立認證會話，跳過測試');
    return;
  }
  
  logInfo(`✅ 認證成功，用戶: ${session.user.email}`);

  // 根據測試類型執行對應的測試
  group(`WebSocket ${testType} Test`, () => {
    switch (testType) {
      case 'connections':
        testConnectionStress(config, session);
        break;
      case 'messaging':
        testMessagingStress(config, session);
        break;
      case 'mixed':
      default:
        testMixedStress(config, session);
        break;
    }
  });

  logGroupEnd(`WebSocket 壓力測試完成`);
  
  // 適當的等待時間
  sleep(1);
}
