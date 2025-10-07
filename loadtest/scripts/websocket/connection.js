/**
 * WebSocket 測試 - 連線 (Connection)
 * 
 * 後端 WebSocket 協議：
 * - 連線 URL: ws://host/ws?token=ACCESS_TOKEN
 * - 訊息格式: { "action": "action_name", "data": { ... } }
 * - 支援的 actions: join_room, leave_room, send_message, ping
 * - 回應 actions: room_joined, room_left, message_sent, new_message, error, pong
 */
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { logWebSocketEvent, logInfo, logError } from '../common/logger.js';

// 自訂量測指標
export const ws_connect_time = new Trend('ws_connect_time');
export const ws_connection_success = new Counter('ws_connection_success');
export const ws_connection_failed = new Counter('ws_connection_failed');
export const ws_messages_received = new Counter('ws_messages_received');
export const ws_messages_sent = new Counter('ws_messages_sent');

/**
 * 建立 WebSocket 連線並執行測試邏輯
 * 
 * @param {string} wsUrl - WebSocket 基礎 URL (例如: ws://localhost:8111/ws)
 * @param {string} token - Access Token
 * @param {function} handler - 連線成功後執行的處理函數
 * @param {number} timeout - 連線超時時間（秒），預設 30 秒
 * @returns {Object} 測試結果
 */
export default function (wsUrl, token, handler, timeout = 30) {
  const connectionId = `${__VU}_${__ITER}_${Date.now()}`;
  logInfo(`🔌 開始 WebSocket 連線: ${connectionId}`);
  
  if (!token) {
    logError('缺少 Access Token，無法建立 WebSocket 連線');
    ws_connection_failed.add(1);
    return { success: false, error: 'Missing token' };
  }
  
  // 構建完整的連線 URL（token 作為查詢參數）
  const fullUrl = `${wsUrl}?token=${token}`;
  logInfo(`連線 URL: ${wsUrl}?token=${token.substring(0, 20)}...`);
  
  try {
    const start = Date.now();
    let connectionEstablished = false;
    let messagesReceived = 0;
    let messagesSent = 0;
    const receivedMessages = [];
    const messageStates = {}; // 追蹤訊息狀態標誌
    
    const res = ws.connect(fullUrl, {}, function (socket) {
      connectionEstablished = true;
      const connectTime = Date.now() - start;
      ws_connect_time.add(connectTime);
      ws_connection_success.add(1);
      
      logWebSocketEvent('connection_established', `連線成功 (${connectTime}ms)`, {
        connection_id: connectionId,
        connect_time: connectTime
      });
      
      // 設置訊息監聽器
      socket.on('message', function (message) {
        messagesReceived++;
        ws_messages_received.add(1);
        
        try {
          const data = JSON.parse(message);
          receivedMessages.push(data);
          
          // 更新狀態標誌
          if (data.action) {
            messageStates[data.action] = true;
            messageStates[`${data.action}_count`] = (messageStates[`${data.action}_count`] || 0) + 1;
            messageStates[`${data.action}_data`] = data.data;
          }
          
          logWebSocketEvent('message_received', `收到訊息 (#${messagesReceived})`, {
            action: data.action,
            connection_id: connectionId,
            total_received: receivedMessages.length
          });
          
          // 根據 action 類型記錄不同的訊息
          switch (data.action) {
            case 'room_joined':
              logInfo(`✅ 收到 room_joined: ${data.data?.message || 'unknown'}`);
              break;
            case 'room_left':
              logInfo(`✅ 收到 room_left: ${data.data?.message || 'unknown'}`);
              break;
            case 'message_sent':
              logInfo(`✅ 收到 message_sent`);
              break;
            case 'new_message':
              logInfo(`📨 收到 new_message: ${data.data?.content || 'empty'}`);
              break;
            case 'pong':
              logInfo(`🏓 收到 pong`);
              break;
            case 'error':
              logError(`❌ WebSocket 錯誤: ${data.data?.message || 'unknown error'}`);
              break;
            default:
              logInfo(`📬 收到訊息: action=${data.action}`);
          }
          
        } catch (e) {
          logWebSocketEvent('message_parse_error', `訊息解析失敗`, {
            error: e.message,
            raw_message: message.substring(0, 100)
          });
        }
      });
      
      // 設置錯誤監聽器
      socket.on('error', function (e) {
        logError(`WebSocket 連線錯誤`, e.error());
        ws_connection_failed.add(1);
      });
      
      // 設置關閉監聽器
      socket.on('close', function () {
        const totalTime = Date.now() - start;
        logWebSocketEvent('connection_closed', `連線關閉`, {
          connection_id: connectionId,
          total_time: totalTime,
          messages_received: messagesReceived,
          messages_sent: messagesSent
        });
      });
      
      // 調用 handler 執行測試邏輯（只發送訊息，不檢查結果）
      if (typeof handler === 'function') {
        logInfo(`執行 WebSocket 測試邏輯: ${connectionId}`);
        try {
          const result = handler(socket);
          
          // 追蹤發送的訊息數量
          messagesSent = result?.messagesSent || 0;
          ws_messages_sent.add(messagesSent);
          
        } catch (e) {
          logError(`測試邏輯執行失敗`, e.message);
        }
      }
      
      // ⭐ 關鍵：等待足夠長的時間讓所有訊息到達
      // 這個 sleep 必須在 handler 返回後執行，確保所有異步訊息都能接收
      logInfo(`⏳ 等待訊息接收完成...`);
      sleep(5);
      
    });
    
    // 檢查連線狀態
    const connectionCheck = check(res, { 
      'WS connection: status is 101': (r) => {
        if (r && r.status === 101) {
          logInfo(`✅ WebSocket 連線狀態檢查通過: ${r.status}`);
          return true;
        } else {
          logError(`❌ WebSocket 連線狀態檢查失敗`, `期望: 101, 實際: ${r?.status || 'undefined'}`);
          ws_connection_failed.add(1);
          return false;
        }
      }
    });
    
    // ⭐ 關鍵：ws.connect() 返回後，socket.on('message') 的回調才會執行
    // 現在 receivedMessages 和 messageStates 已經有正確的資料了
    
    const duration = Date.now() - start;
    
    // 記錄最終統計
    logInfo(`📊 最終統計 - 收到 ${receivedMessages.length} 條訊息，發送 ${messagesSent} 條訊息`);
    const stateKeys = Object.keys(messageStates).filter(k => !k.endsWith('_count') && !k.endsWith('_data'));
    if (stateKeys.length > 0) {
      logInfo(`📊 訊息狀態: ${stateKeys.join(', ')}`);
    }
    
    // 更新指標
    if (connectionCheck && connectionEstablished) {
      ws_connection_success.add(1);
      ws_connection_duration.add(duration);
    }
    
    // 返回測試結果（包含完整的訊息狀態）
    return {
      success: connectionCheck && connectionEstablished,
      messagesReceived,
      messagesSent,
      receivedMessages,
      messageStates, // 這時已經有正確的資料了
      connectionTime: duration
    };
    
  } catch (e) {
    const errorMsg = `WebSocket 連線異常: ${e.message}`;
    logError(errorMsg, {
      connection_id: connectionId,
      url: wsUrl,
      stack: e.stack
    });
    ws_connection_failed.add(1);
    
    return {
      success: false,
      error: e.message
    };
  }
}

/**
 * 建立簡單的 WebSocket 連線測試（只連線不發送訊息）
 * 
 * @param {string} wsUrl - WebSocket URL
 * @param {string} token - Access Token
 * @returns {boolean} 連線是否成功
 */
export function testConnection(wsUrl, token) {
  logInfo('測試 WebSocket 連線能力');
  
  const result = ws.connect(`${wsUrl}?token=${token}`, {}, function (socket) {
    logInfo('✅ WebSocket 連線成功建立');
    sleep(1);
  });
  
  return check(result, {
    'WS simple connection test: status is 101': (r) => r.status === 101
  });
}
