/**
 * API 測試 - 好友 (Friends)
 */
import { group, check } from 'k6';
import http from 'k6/http';
import { applyCsrf } from '../common/csrf.js';
import { logHttpResponse, logGroupStart, logGroupEnd, logInfo } from '../common/logger.js';

export default function (baseUrl, session, targetUsername, friendId) {
  if (!session) return;

  const groupStartTime = Date.now();
  logGroupStart('Friend Management APIs');

  group('API - Friends', function () {
    group('Get Friend List', function () {
      // 根據 API.md: GET /friends
      logInfo('取得好友列表');
      const headers = {
        ...session.headers,
      };
      const res = http.get(`${baseUrl}/friends`, { headers });
      
      logHttpResponse('GET /friends', res, { expectedStatus: 200 });
      
      check(res, { 
        'Get Friend List: status is 200': (r) => r.status === 200,
        'Get Friend List: response has success status': (r) => r.json('status') === 'success'
      });
    });

    group('Get Pending Requests', function () {
      // 根據 API.md: GET /friends/pending
      logInfo('取得待處理好友請求');
      const headers = {
        ...session.headers,
      };
      const res = http.get(`${baseUrl}/friends/pending`, { headers });
      
      logHttpResponse('GET /friends/pending', res, { expectedStatus: 200 });
      
      check(res, { 
        'Get Pending Requests: status is 200': (r) => r.status === 200,
        'Get Pending Requests: response has success status': (r) => r.json('status') === 'success'
      });
    });

    group('Get Blocked Users', function () {
      // 根據 API.md: GET /friends/blocked
      logInfo('取得封鎖列表');
      const headers = {
        ...session.headers,
      };
      const res = http.get(`${baseUrl}/friends/blocked`, { headers });
      
      logHttpResponse('GET /friends/blocked', res, { expectedStatus: 200 });
      
      check(res, { 
        'Get Blocked Users: status is 200': (r) => r.status === 200,
        'Get Blocked Users: response has success status': (r) => r.json('status') === 'success'
      });
    });

    if (targetUsername) {
      group('Send Friend Request', function () {
        // 根據 API.md: POST /friends/requests
        logInfo(`發送好友邀請給: ${targetUsername}`);
        const url = `${baseUrl}/friends/requests`;
        const payload = JSON.stringify({
          username: targetUsername
        });
        const headers = {
          'Content-Type': 'application/json',
          ...session.headers,
          ...applyCsrf(url, {}, session.csrfToken),
        };
        const res = http.post(url, payload, { headers });
        
        logHttpResponse('POST /friends/requests', res, { expectedStatus: [200, 400, 403] });
        
        check(res, { 
          'Send Friend Request: status is 200, 400 or 403': (r) => r.status === 200 || r.status === 400 || r.status === 403,
          'Send Friend Request: response has status field': (r) => r.json('status') !== undefined
        });
        
        if (res.status === 200) {
          logInfo(`✅ 好友邀請發送成功`);
        } else if (res.status === 403) {
          logInfo(`ℹ️  好友邀請被拒絕 (CSRF 驗證或權限問題)`);
        } else if (res.status === 400) {
          logInfo(`ℹ️  好友邀請處理 (可能已存在或其他原因)`);
        }
      });
    }

    if (friendId) {
      group('Update Friend Status', function() {
        // 根據 API.md: PUT /friends/requests/{request_id}/accept
        logInfo(`接受好友請求: ${friendId}`);
        const url = `${baseUrl}/friends/requests/${friendId}/accept`;
        const headers = {
          ...session.headers,
          ...applyCsrf(url, {}, session.csrfToken),
        };
        const res = http.put(url, null, { headers });
        
        logHttpResponse('PUT /friends/requests/{id}/accept', res, { expectedStatus: [200, 400, 403] });
        
        check(res, { 
          'Update Friend Status: status is 200, 400 or 403': (r) => r.status === 200 || r.status === 400 || r.status === 403,
          'Update Friend Status: response has status field': (r) => r.json('status') !== undefined
        });
        
        if (res.status === 200) {
          logInfo(`✅ 好友狀態更新成功`);
        } else if (res.status === 403) {
          logInfo(`ℹ️  好友狀態更新被拒絕 (CSRF 驗證或權限問題)`);
        }
      });
    }
  });
  
  logGroupEnd('Friend Management APIs', groupStartTime);
}
