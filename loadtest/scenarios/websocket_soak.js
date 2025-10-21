/**
 * WebSocket 浸泡測試場景 (Soak Test)
 * 
 * 目的：測試系統長時間運行的穩定性和記憶體洩漏問題
 * 
 * 測試特點：
 * 1. 中等負載
 * 2. 長時間運行（1-2 小時）
 * 3. 持續監控系統指標
 * 4. 檢測記憶體洩漏和性能衰退
 * 
 * 使用方法：
 * k6 run run.js --env SCENARIO=websocket_soak --env SOAK_DURATION=3600
 */

import { check, sleep, group } from 'k6';
import ws from 'k6/ws';
import { Counter, Trend, Rate, Gauge } from 'k6/metrics';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { logInfo, logError, logSuccess } from '../scripts/common/logger.js';

// 浸泡測試專用指標
const soakConnectionUptime = new Trend('soak_connection_uptime');
const soakMessageLatency = new Trend('soak_message_latency');
const soakConnectionStability = new Rate('soak_connection_stability');
const soakMemoryIndicator = new Gauge('soak_memory_indicator');

export default function (config) {
  const vuNumber = __VU;
  const soakDuration = parseInt(__ENV.SOAK_DURATION || '3600'); // 預設 1 小時
  
  logInfo(`🔥 浸泡測試 VU ${vuNumber} - 目標運行時間: ${soakDuration}秒`);

  // 取得認證會話
  const session = getAuthenticatedSession(`${config.BASE_URL}${config.API_PREFIX}`);
  
  if (!session) {
    logError('認證失敗');
    return;
  }

  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const soakRoomId = `soak_room_${vuNumber % 10}`; // 10 個房間輪流使用

  group('Soak Test: 長時間穩定性測試', () => {
    const testStartTime = Date.now();
    
    const response = ws.connect(wsUrl, {
      headers: {
        'Authorization': `Bearer ${session.token}`,
      },
      tags: { test_type: 'soak' },
    }, function (socket) {
      logSuccess(`VU ${vuNumber} 浸泡測試連線成功`);

      let messageCount = 0;
      let errorCount = 0;
      let isConnected = true;

      socket.on('open', () => {
        logInfo(`VU ${vuNumber} 開始浸泡測試，目標: ${soakDuration}秒`);
        
        // 加入房間
        socket.send(JSON.stringify({
          action: 'join_room',
          room_id: soakRoomId
        }));

        // 開始定期發送訊息（模擬真實用戶行為）
        startSoakMessaging(socket);
      });

      socket.on('message', (data) => {
        try {
          const receiveTime = Date.now();
          const message = JSON.parse(data);
          
          if (message.action === 'new_message' && message.timestamp) {
            const sentTime = new Date(message.timestamp).getTime();
            const latency = receiveTime - sentTime;
            soakMessageLatency.add(latency);
            
            // 監控延遲是否增長（可能表示記憶體洩漏或性能衰退）
            if (latency > 1000) {
              logError(`VU ${vuNumber} 訊息延遲過高: ${latency}ms`);
              soakMemoryIndicator.add(latency);
            }
          }
          
          if (message.action === 'error') {
            errorCount++;
            soakConnectionStability.add(0);
          }
        } catch (e) {
          errorCount++;
          logError(`VU ${vuNumber} 解析訊息失敗: ${e.message}`);
        }
      });

      socket.on('error', (e) => {
        isConnected = false;
        errorCount++;
        logError(`VU ${vuNumber} WebSocket 錯誤: ${e.error()}`);
        soakConnectionStability.add(0);
      });

      socket.on('close', () => {
        const uptime = Date.now() - testStartTime;
        soakConnectionUptime.add(uptime);
        
        const stabilityRate = errorCount === 0 ? 1 : Math.max(0, 1 - (errorCount / messageCount));
        soakConnectionStability.add(stabilityRate);
        
        logInfo(`VU ${vuNumber} 浸泡測試結束`);
        logInfo(`  運行時間: ${uptime}ms (${(uptime / 1000).toFixed(2)}秒)`);
        logInfo(`  發送訊息: ${messageCount}`);
        logInfo(`  錯誤次數: ${errorCount}`);
        logInfo(`  穩定率: ${(stabilityRate * 100).toFixed(2)}%`);
      });

      function startSoakMessaging(socket) {
        let messageInterval = 0;
        const maxMessages = Math.floor(soakDuration / 30); // 每 30 秒發送一條訊息
        
        const sendNextMessage = () => {
          if (messageInterval >= maxMessages || socket.readyState !== 1) {
            // 測試結束，離開房間並關閉連線
            socket.send(JSON.stringify({
              action: 'leave_room',
              room_id: soakRoomId
            }));
            socket.setTimeout(() => socket.close(), 1000);
            return;
          }

          if (socket.readyState === 1) {
            const timestamp = new Date().toISOString();
            socket.send(JSON.stringify({
              action: 'send_message',
              room_id: soakRoomId,
              content: `浸泡測試訊息 #${messageCount + 1} from VU ${vuNumber}`,
              message_type: 'text',
              timestamp: timestamp
            }));
            
            messageCount++;
            
            // 每 100 條訊息記錄一次
            if (messageCount % 100 === 0) {
              const elapsedTime = (Date.now() - testStartTime) / 1000;
              logInfo(`VU ${vuNumber} 已運行 ${elapsedTime.toFixed(0)}秒，發送 ${messageCount} 條訊息`);
            }
          }

          messageInterval++;
          
          // 每 30 秒 ± 5 秒發送一條訊息（模擬真實用戶）
          const nextDelay = (30 + (Math.random() * 10 - 5)) * 1000;
          socket.setTimeout(sendNextMessage, nextDelay);
        };

        // 開始發送訊息
        sendNextMessage();
      }
    });

    check(response, {
      '浸泡測試 WebSocket 連線建立': (r) => r && r.status === 101,
    });
  });

  // 輕微的隨機延遲
  sleep(1 + Math.random());
}
