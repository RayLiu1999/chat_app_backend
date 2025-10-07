/**
 * WebSocket 大量連線測試
 * 用於測試同時大量 WebSocket 連線的性能
 */
import { group, check } from 'k6';
import ws from 'k6/ws';
import { Counter, Rate, Trend } from 'k6/metrics';
import { logWebSocketEvent, logInfo, logError } from '../common/logger.js';
import { getAuthenticatedSession } from '../common/auth.js';
import * as config from '../../config.js';

// WebSocket 專用指標
export const ws_concurrent_connections = new Counter('ws_concurrent_connections');
export const ws_connection_success_rate = new Rate('ws_connection_success_rate');
export const ws_message_send_time = new Trend('ws_message_send_time');
export const ws_room_join_time = new Trend('ws_room_join_time');

// 測試配置
export const options = {
  scenarios: {
    // 漸進式增加連線數
    websocket_ramp_up: {
      executor: 'ramping-vus',
      startVUs: 1,
      stages: [
        { duration: '30s', target: 10 },   // 30秒內增加到10個連線
        { duration: '1m', target: 50 },    // 1分鐘內增加到50個連線
        { duration: '1m', target: 100 },   // 1分鐘內增加到100個連線
        { duration: '2m', target: 100 },   // 維持100個連線2分鐘
        { duration: '30s', target: 0 },    // 30秒內降到0
      ],
    },
  },
  thresholds: {
    'ws_connection_success_rate': ['rate>0.95'], // 95% 連線成功率
    'ws_connecting': ['p(95)<1000'],              // 95% 連線時間 < 1秒
    'ws_message_send_time': ['p(95)<500'],       // 95% 訊息發送時間 < 500ms
    'ws_room_join_time': ['p(95)<2000'],         // 95% 房間加入時間 < 2秒
  },
};

export default function () {
  const baseUrl = `${config.TEST_CONFIG.BASE_URL}${config.TEST_CONFIG.API_PREFIX}`;
  const wsUrl = config.TEST_CONFIG.WS_URL;
  
  group('WebSocket 大量連線測試', function () {
    // 取得認證 session
    const session = getAuthenticatedSession(baseUrl);
    if (!session || !session.token) {
      logError('無法取得認證 token，跳過 WebSocket 測試');
      return;
    }
    
    const vuId = __VU;
    const iterationId = __ITER;
    const connectionId = `VU${vuId}_Iter${iterationId}_${Date.now()}`;
    
    logInfo(`🚀 VU ${vuId} 開始 WebSocket 連線測試 (${connectionId})`);
    
    try {
      const connectionStart = Date.now();
      let connectionEstablished = false;
      let roomJoined = false;
      let messagesSent = 0;
      
      const res = ws.connect(`${wsUrl}?token=${session.token}`, {
        timeout: '10s',
      }, function (socket) {
        connectionEstablished = true;
        ws_concurrent_connections.add(1);
        ws_connection_success_rate.add(1);
        
        const connectTime = Date.now() - connectionStart;
        logWebSocketEvent('mass_connection_established', 
          `大量連線建立成功 (${connectTime}ms)`, {
            vu_id: vuId,
            connection_id: connectionId,
            connect_time: connectTime
          });
        
        // 訊息接收處理
        socket.on('message', function (message) {
          try {
            const data = JSON.parse(message);
            
            // 處理不同類型的訊息
            switch (data.type) {
              case 'room_joined':
                if (!roomJoined) {
                  roomJoined = true;
                  const joinTime = Date.now() - connectionStart;
                  ws_room_join_time.add(joinTime);
                  logWebSocketEvent('room_join_success', 
                    `房間加入成功 (${joinTime}ms)`, {
                      room_id: data.data?.room_id,
                      connection_id: connectionId
                    });
                }
                break;
                
              case 'message_sent':
                const sendTime = Date.now() - connectionStart;
                ws_message_send_time.add(sendTime);
                logWebSocketEvent('message_send_confirmed', 
                  `訊息發送確認 (${sendTime}ms)`, {
                    message_id: data.data?.message_id,
                    connection_id: connectionId
                  });
                break;
                
              case 'error':
                logError(`WebSocket 服務器錯誤`, {
                  error: data.message,
                  connection_id: connectionId
                });
                break;
                
              case 'user_joined':
                logWebSocketEvent('user_activity', '用戶加入房間', {
                  user_id: data.data?.user_id,
                  room_id: data.data?.room_id
                });
                break;
                
              case 'user_left':
                logWebSocketEvent('user_activity', '用戶離開房間', {
                  user_id: data.data?.user_id,
                  room_id: data.data?.room_id
                });
                break;
                
              default:
                logWebSocketEvent('unknown_message', `未知訊息類型: ${data.type}`, data);
            }
            
          } catch (e) {
            logError('訊息解析失敗', {
              error: e.message,
              raw_message: message.substring(0, 200),
              connection_id: connectionId
            });
          }
        });
        
        // 錯誤處理
        socket.on('error', function (e) {
          logError('WebSocket 連線錯誤', {
            error: e.error(),
            connection_id: connectionId
          });
        });
        
        // 連線關閉處理
        socket.on('close', function () {
          const totalTime = Date.now() - connectionStart;
          logWebSocketEvent('mass_connection_closed', `大量連線關閉`, {
            connection_id: connectionId,
            total_time: totalTime,
            messages_sent: messagesSent,
            room_joined: roomJoined
          });
        });
        
        // 執行測試流程
        setTimeout(() => {
          // 1. 加入測試房間
          const testRoomId = `load_test_room_${vuId % 5}`; // 分散到5個房間
          const joinMessage = JSON.stringify({
            type: 'join_room',
            data: {
              room_id: testRoomId,
              room_type: 'dm'
            }
          });
          
          logWebSocketEvent('sending_join_room', `加入房間: ${testRoomId}`, {
            connection_id: connectionId
          });
          socket.send(joinMessage);
          
        }, 100); // 連線後等待100ms
        
        setTimeout(() => {
          // 2. 發送測試訊息
          const testMessage = JSON.stringify({
            type: 'send_message',
            data: {
              room_id: `load_test_room_${vuId % 5}`,
              room_type: 'dm',
              content: `Load test message from VU ${vuId} at ${new Date().toISOString()}`
            }
          });
          
          messagesSent++;
          logWebSocketEvent('sending_message', `發送訊息 #${messagesSent}`, {
            connection_id: connectionId
          });
          socket.send(testMessage);
          
        }, 500); // 加入房間後等待500ms發送訊息
        
        setTimeout(() => {
          // 3. 發送第二條訊息
          const testMessage2 = JSON.stringify({
            type: 'send_message',
            data: {
              room_id: `load_test_room_${vuId % 5}`,
              room_type: 'dm',
              content: `Second message from VU ${vuId}`
            }
          });
          
          messagesSent++;
          socket.send(testMessage2);
          
        }, 1000); // 1秒後發送第二條訊息
        
        setTimeout(() => {
          // 4. 離開房間
          const leaveMessage = JSON.stringify({
            type: 'leave_room',
            data: {
              room_id: `load_test_room_${vuId % 5}`,
              room_type: 'dm'
            }
          });
          
          logWebSocketEvent('sending_leave_room', '離開房間', {
            connection_id: connectionId
          });
          socket.send(leaveMessage);
          
        }, 1500); // 1.5秒後離開房間
        
        // 5. 保持連線一段時間後關閉
        setTimeout(() => {
          logWebSocketEvent('closing_connection', '主動關閉連線', {
            connection_id: connectionId
          });
          socket.close();
        }, 3000); // 3秒後關閉連線
      });
      
      // 檢查連線建立
      check(res, {
        'WebSocket 大量連線: 狀態為 101': (r) => {
          const success = r && r.status === 101;
          if (!success) {
            ws_connection_success_rate.add(0);
            logError('WebSocket 連線建立失敗', {
              status: r?.status,
              connection_id: connectionId
            });
          }
          return success;
        }
      });
      
    } catch (e) {
      ws_connection_success_rate.add(0);
      logError('WebSocket 連線異常', {
        error: e.message,
        connection_id: connectionId,
        vu_id: vuId
      });
    }
  });
  
  // 隨機等待時間，避免所有 VU 同時操作
  const randomDelay = Math.random() * 1000 + 500; // 500-1500ms
  logInfo(`VU ${vuId} 等待 ${randomDelay.toFixed(0)}ms 後結束迭代`);
}

// 測試摘要
export function handleSummary(data) {
  console.log('\n🔌 WebSocket 大量連線測試完成！');
  console.log('=' .repeat(60));
  console.log(`📊 連線總數: ${data.metrics.ws_concurrent_connections?.values?.count || 0}`);
  console.log(`✅ 連線成功率: ${((data.metrics.ws_connection_success_rate?.values?.rate || 0) * 100).toFixed(2)}%`);
  console.log(`⏱️  平均連線時間: ${(data.metrics.ws_connecting?.values?.avg || 0).toFixed(2)}ms`);
  console.log(`📈 95% 連線時間: ${(data.metrics.ws_connecting?.values?.['p(95)'] || 0).toFixed(2)}ms`);
  
  if (data.metrics.ws_room_join_time) {
    console.log(`🏠 平均房間加入時間: ${data.metrics.ws_room_join_time.values.avg.toFixed(2)}ms`);
    console.log(`🏠 95% 房間加入時間: ${data.metrics.ws_room_join_time.values['p(95)'].toFixed(2)}ms`);
  }
  
  if (data.metrics.ws_message_send_time) {
    console.log(`💬 平均訊息發送時間: ${data.metrics.ws_message_send_time.values.avg.toFixed(2)}ms`);
    console.log(`💬 95% 訊息發送時間: ${data.metrics.ws_message_send_time.values['p(95)'].toFixed(2)}ms`);
  }
  
  console.log('=' .repeat(60));
  
  return {
    stdout: 'WebSocket 大量連線測試已完成\n'
  };
}
