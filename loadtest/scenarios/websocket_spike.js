/**
 * WebSocket 峰值測試場景 (Spike Test)
 * 
 * 目的：測試系統在突然增加大量 WebSocket 連線時的表現
 * 
 * 測試場景：
 * 1. 突然湧入大量用戶連線
 * 2. 所有用戶同時發送訊息
 * 3. 測試系統恢復能力
 * 
 * 使用方法：
 * k6 run run.js --env SCENARIO=websocket_spike
 */

import { check, sleep, group } from 'k6';
import ws from 'k6/ws';
import { Counter, Trend, Rate } from 'k6/metrics';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { logInfo, logError, logSuccess } from '../scripts/common/logger.js';

// 自定義指標
const spikeConnectionSuccess = new Rate('spike_connection_success');
const spikeMessageDelivery = new Rate('spike_message_delivery');
const spikeSystemRecovery = new Rate('spike_system_recovery');

export default function (config) {
  const vuNumber = __VU;
  const iteration = __ITER;
  
  logInfo(`🚀 峰值測試 VU ${vuNumber} - 迭代 ${iteration}`);

  // 取得認證會話
  const session = getAuthenticatedSession(`${config.BASE_URL}${config.API_PREFIX}`);
  
  if (!session) {
    logError('認證失敗');
    return;
  }

  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const spikeRoomId = 'spike_test_room'; // 所有用戶使用同一個房間

  group('Spike Test: 突發連線與訊息', () => {
    const connectionStart = Date.now();
    
    const response = ws.connect(wsUrl, {
      headers: {
        'Authorization': `Bearer ${session.token}`,
      },
      tags: { test_type: 'spike' },
    }, function (socket) {
      const connectionTime = Date.now() - connectionStart;
      
      const connectionSuccess = check(null, {
        'Spike 連線成功': () => connectionTime < 3000, // 3 秒內連線成功
      });
      
      spikeConnectionSuccess.add(connectionSuccess);
      
      if (connectionSuccess) {
        logSuccess(`VU ${vuNumber} 連線成功`, 101, connectionTime);
      } else {
        logError(`VU ${vuNumber} 連線過慢: ${connectionTime}ms`);
      }

      let joinedRoom = false;
      let messageSent = false;
      let messageReceived = false;

      socket.on('open', () => {
        // 立即加入房間
        socket.send(JSON.stringify({
          action: 'join_room',
          room_id: spikeRoomId
        }));
      });

      socket.on('message', (data) => {
        try {
          const message = JSON.parse(data);
          
          if (message.action === 'status' && !joinedRoom) {
            joinedRoom = true;
            logInfo(`VU ${vuNumber} 加入峰值測試房間`);
            
            // 加入房間後立即發送訊息（模擬突發訊息）
            const sendStart = Date.now();
            socket.send(JSON.stringify({
              action: 'send_message',
              room_id: spikeRoomId,
              content: `峰值測試訊息 from VU ${vuNumber}`,
              message_type: 'text'
            }));
            messageSent = true;
            
            const sendTime = Date.now() - sendStart;
            logInfo(`VU ${vuNumber} 發送訊息耗時: ${sendTime}ms`);
          }
          
          if (message.action === 'new_message') {
            messageReceived = true;
            spikeMessageDelivery.add(1);
            logInfo(`VU ${vuNumber} 收到廣播訊息`);
          }
        } catch (e) {
          logError(`VU ${vuNumber} 解析訊息失敗: ${e.message}`);
        }
      });

      socket.on('error', (e) => {
        logError(`VU ${vuNumber} WebSocket 錯誤: ${e.error()}`);
        spikeSystemRecovery.add(0);
      });

      socket.on('close', () => {
        const recovered = joinedRoom && messageSent;
        spikeSystemRecovery.add(recovered ? 1 : 0);
        
        logInfo(`VU ${vuNumber} 測試完成 - 加入: ${joinedRoom}, 發送: ${messageSent}, 接收: ${messageReceived}`);
      });

      // 保持連線 10 秒後關閉
      socket.setTimeout(() => {
        if (socket.readyState === 1) {
          socket.send(JSON.stringify({
            action: 'leave_room',
            room_id: spikeRoomId
          }));
        }
        socket.setTimeout(() => socket.close(), 500);
      }, 10000);
    });

    check(response, {
      'Spike 測試 WebSocket 連線建立': (r) => r && r.status === 101,
    });
  });

  sleep(1);
}
