/**
 * API 測試 - 伺服器 (Servers)
 */
import { group, check } from 'k6';
import http from 'k6/http';
import { FormData } from 'https://jslib.k6.io/formdata/0.0.2/index.js';
import { randomString } from '../common/utils.js';
import { applyCsrf } from '../common/csrf.js';
import { logHttpResponse, logGroupStart, logGroupEnd, logInfo } from '../common/logger.js';

export default function (baseUrl, session) {
  if (!session) return null;

  const groupStartTime = Date.now();
  logGroupStart('Server Management APIs');

  let serverId;
  let channelId;

  group('API - Servers', function () {

    group('Get Server List', function () {
      logInfo('取得伺服器列表');
      const headers = {
        ...session.headers,
        "Origin": "http://localhost:3000",
      };
      const res = http.get(`${baseUrl}/servers`, { headers });

      logHttpResponse('GET /servers', res);

      check(res, { 
        'Get Server List: status is 200': (r) => r.status === 200,
        'Get Server List: response has success status': (r) => r.json('status') === 'success'
      });
    });

    group('Search Servers', function () {
      logInfo('搜尋伺服器: query=test');
      const headers = {
        ...session.headers,
        "Origin": "http://localhost:3000",
      };
      const res = http.get(`${baseUrl}/servers/search?query=test&page=1&limit=10`, { headers });
      
      logHttpResponse('GET /servers/search', res);
      
      check(res, { 
        'Search Servers: status is 200': (r) => r.status === 200,
        'Search Servers: response has success status': (r) => r.json('status') === 'success'
      });
    });

    group('Create Server', function () {
      const serverName = `Test Server ${randomString(6)}`;
      logInfo(`創建伺服器: ${serverName}`);
      
      const url = `${baseUrl}/servers`;
      const formData = new FormData();
      formData.append('name', serverName);
      formData.append('description', 'Test server created by k6 load test');
      formData.append('is_public', 'false');

      const headers = {
        ...session.headers,
        ...applyCsrf(url, {}),
        "Content-Type": `multipart/form-data; boundary=${formData.boundary}`,
        "Origin": "http://localhost:3000",
      };
      const res = http.post(url, formData.body(), { headers });
      let body;
      try {
        body = res.json();
      } catch (e) {
        body = null;
      }
    
      logHttpResponse('POST /servers', res, { expectedStatus: 200 });
    
      if (check(res, {
        'Create Server: status is 200': (r) => r.status === 200,
        'Create Server: response has success status': () => body && body.status === 'success',
        'Create Server: has server data': () => body && body.data && body.data.id !== undefined
      })) {
        serverId = body.data.id;
        logInfo(`✅ 伺服器創建成功，ID: ${serverId}`);
      } else {
        logInfo(`❌ 伺服器創建失敗: ${res.status} ${res.body}`);
      }
    });

    if (serverId) {
      group('Get Server Details', function () {
        logInfo(`取得伺服器詳細資料: ${serverId}`);
        const headers = {
          ...session.headers,
          "Origin": "http://localhost:3000",
        };
        const res = http.get(`${baseUrl}/servers/${serverId}`, { headers });
        
        logHttpResponse('GET /servers/{id}', res, { expectedStatus: 200 });
        
        check(res, { 
          'Get Server Details: status is 200': (r) => r.status === 200,
          'Get Server Details: response has success status': (r) => r.json('status') === 'success'
        });
      });

      group('Get Server Detail with Members', function () {
        logInfo(`取得伺服器詳細資料含成員: ${serverId}`);
        const headers = {
          ...session.headers,
          "Origin": "http://localhost:3000",
        };
        const res = http.get(`${baseUrl}/servers/${serverId}/detail`, { headers });
        
        logHttpResponse('GET /servers/{id}/detail', res, { expectedStatus: 200 });
        
        check(res, { 
          'Get Server Detail: status is 200': (r) => r.status === 200,
          'Get Server Detail: response has success status': (r) => r.json('status') === 'success'
        });
      });

      group('Get Server Channels', function () {
        logInfo(`取得伺服器頻道列表: ${serverId}`);
        const headers = {
          ...session.headers,
          "Origin": "http://localhost:3000",
        };
        const res = http.get(`${baseUrl}/servers/${serverId}/channels`, { headers });
        
        logHttpResponse('GET /servers/{id}/channels', res, { expectedStatus: 200 });
        
        check(res, { 
          'Get Server Channels: status is 200': (r) => r.status === 200,
          'Get Server Channels: response has success status': (r) => r.json('status') === 'success'
        });
      });

      group('Create Channel', function() {
        const channelName = `general-${randomString(4)}`;
        logInfo(`在伺服器 ${serverId} 創建頻道: ${channelName}`);
        
        const url = `${baseUrl}/servers/${serverId}/channels`;
        const payload = JSON.stringify({
          name: channelName,
          type: 'text'
        });
        const headers = {
          'Content-Type': 'application/json',
          ...session.headers,
          ...applyCsrf(url, {}),
          'Origin': 'http://localhost:3000',
        };

        const res = http.post(url, payload, { headers });
        
        logHttpResponse('POST /servers/{id}/channels', res, { expectedStatus: [200, 201] });
        
        if (check(res, { 
          'Create Channel: status is 200 or 201': (r) => r.status === 200 || r.status === 201,
          'Create Channel: response has success status': (r) => r.json('status') === 'success',
          'Create Channel: has channel data': (r) => r.json('data.id') !== undefined
        })) {
          channelId = res.json('data.id');
          logInfo(`✅ 頻道 ${channelName} 創建成功，ID: ${channelId}`);
        }
      });

      group('Update Server', function () {
        const newName = `Updated Test Server ${randomString(4)}`;
        logInfo(`更新伺服器 ${serverId}: ${newName}`);
        
        const url = `${baseUrl}/servers/${serverId}`;
        const payload = JSON.stringify({
          name: newName,
          description: `Updated description`,
          is_public: true
        });
        const headers = {
          'Content-Type': 'application/json',
          ...session.headers,
          ...applyCsrf(url, {}, session.csrfToken),
          'Origin': 'http://localhost:3000',
        };
        const res = http.put(url, payload, { headers });
        
        logHttpResponse('PUT /servers/{id}', res, { expectedStatus: 200 });
        
        check(res, { 
          'Update Server: status is 200': (r) => r.status === 200,
          'Update Server: response has success status': (r) => r.json('status') === 'success'
        });
        
        if (res.status === 200) {
          logInfo(`✅ 伺服器更新成功: ${newName}`);
        }
      });
    }
  });
  
  logGroupEnd('Server Management APIs', groupStartTime);
  
  // 返回創建的伺服器和頻道 ID，供後續測試使用
  return {
    serverId,
    channelId
  };
}
