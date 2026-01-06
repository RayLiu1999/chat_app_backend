/**
 * API 測試 - 認證 (Auth)
 */
import { group, check } from 'k6';
import http from 'k6/http';
import { registerUser, login } from '../common/auth.js';
import { randomItem } from '../common/utils.js';
import { logHttpResponse, logGroupStart, logGroupEnd, logInfo } from '../common/logger.js';
import { SharedArray } from 'k6/data';

// 讀取測試用戶資料
const users = new SharedArray('users', function () {
  try {
    return JSON.parse(open('../../data/users.json'));
  } catch (e) {
    return [
      { username: 'testuser', email: 'test@example.com', password: 'password123' },
    ];
  }
});

export default function (baseUrl) {
  const groupStartTime = Date.now();
  logGroupStart('Authentication APIs');
  
  group('API - Authentication', function () {
    const user = randomItem(users);
    logInfo(`測試用戶: ${user.email}`);

    // 註冊
    group('Register', function () {
      registerUser(baseUrl, user);
    });

    // 登入
    group('Login', function () {
      login(baseUrl, { email: user.email, password: user.password });
    });

    // 健康檢查
    group('Health Check', function () {
      logInfo('執行系統健康檢查');
      const res = http.get(`${baseUrl}/health`);
      
      let body;
      try {
        body = res.json();
      } catch (e) {
        body = null;
      }
      
      // 記錄 HTTP 回應
      logHttpResponse('GET /health', res, { expectedStatus: 200 });
      
      check(res, { 
        'Health check: status is 200': (r) => r.status === 200,
        'Health check: has status field': () => body && body.status !== undefined,
        'Health check: system status is ok': () => body && body.data && body.data.status === 'ok'
      });
    });
  });
  
  logGroupEnd('Authentication APIs', groupStartTime);
}
