/**
 * 即時日誌展示測試
 * 這個腳本專門用來測試和展示即時日誌功能
 * 
 * 使用方法:
 * k6 run loadtest/demo_logging.js --env VERBOSE=1
 */
import { group } from 'k6';
import http from 'k6/http';
import { logHttpResponse, logGroupStart, logGroupEnd, logInfo, logError, logWebSocketEvent } from './scripts/common/logger.js';
import { applyCsrf } from './scripts/common/csrf.js';
import * as config from './config.js';

export const options = {
  vus: 2, // 2個虛擬用戶
  duration: '30s', // 運行30秒
  thresholds: {
    http_req_duration: ['p(95)<2000'], // 95%的請求應在2秒內完成
  },
};

export default function () {
  const baseUrl = `${config.TEST_CONFIG.BASE_URL}${config.TEST_CONFIG.API_PREFIX}`;
  const groupStartTime = Date.now();
  
  logGroupStart('即時日誌功能展示');
  
  group('Health Check 展示', function () {
    logInfo('開始執行健康檢查');
    
    const res = http.get(`${baseUrl}/health`);
    logHttpResponse('GET /health', res, { expectedStatus: 200 });
    
    if (res.status === 200) {
      try {
        const data = res.json();
        logInfo('健康檢查回應解析成功', data);
      } catch (e) {
        logError('健康檢查回應解析失敗', e.message);
      }
    }
  });
  
  group('模擬登入測試', function () {
    logInfo('嘗試模擬登入請求');
    
    const payload = JSON.stringify({
      email: 'demo@example.com',
      password: 'demo123'
    });
    
    const headers = applyCsrf(`${baseUrl}/login`, { 'Content-Type': 'application/json' });
    const res = http.post(`${baseUrl}/login`, payload, { headers });
    
    logHttpResponse('POST /login', res, { expectedStatus: [200, 400, 401] });
    
    if (res.status === 401) {
      logInfo('預期的未授權回應 - 這是正常的，因為使用了測試憑證');
    }
  });
  
  group('WebSocket 模擬測試', function () {
    logWebSocketEvent('connection_attempt', '嘗試 WebSocket 連線');
    logWebSocketEvent('message_sent', '發送測試訊息', { 
      type: 'send_message', 
      room_id: 'demo_room' 
    });
    logWebSocketEvent('connection_closed', '關閉 WebSocket 連線');
  });
  
  group('錯誤情況展示', function () {
    logInfo('測試不存在的端點');
    
    const res = http.get(`${baseUrl}/nonexistent-endpoint`);
    logHttpResponse('GET /nonexistent-endpoint', res, { expectedStatus: 404 });
    
    if (res.status === 404) {
      logInfo('✅ 正確處理了 404 錯誤');
    }
  });
  
  logGroupEnd('即時日誌功能展示', groupStartTime);
  
  // 在迭代間隨機暫停 1-3 秒
  const sleepTime = Math.random() * 2000 + 1000;
  logInfo(`迭代完成，暫停 ${sleepTime.toFixed(0)}ms`);
  // 使用簡單的延遲而不是 k6 的 sleep，因為這裡只是演示
}

export function handleSummary(data) {
  console.log('\n🎯 即時日誌展示測試完成！');
  console.log('=====================================================');
  console.log(`📊 總請求數: ${data.metrics.http_reqs.values.count}`);
  console.log(`✅ 檢查通過率: ${data.metrics.checks.values.passes}/${data.metrics.checks.values.total}`);
  console.log(`⏱️  平均回應時間: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms`);
  console.log(`📈 95% 回應時間: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms`);
  console.log('=====================================================');
  
  return {
    stdout: '即時日誌展示測試已完成\n'
  };
}
