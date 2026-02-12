/**
 * API 測試 - 聊天 (Chat)
 * 
 * 使用兩個真實用戶測試私聊功能：
 * - user1: 當前登入的用戶 (session)
 * - user2: 從 users.json 獲取，通過測試 API 取得 ID
 * 
 * 新增測試 API：GET /test/user?username=xxx
 * - 不需要認證
 * - 直接通過 username 取得用戶資訊和 ID
 * - 避免登入第二個用戶造成 Cookie 衝突
 */
import { group, check } from 'k6';
import http from 'k6/http';
import { applyCsrf } from '../common/csrf.js';
import { registerUser } from '../common/auth.js';
import { logHttpResponse, logGroupStart, logGroupEnd, logInfo, logError } from '../common/logger.js';
import { SharedArray } from 'k6/data';
import { TEST_CONFIG } from '../../config.js';

// 載入測試用戶
const testUsers = new SharedArray('testUsers', function () {
  try {
    return JSON.parse(open('../../data/users.json'));
  } catch (e) {
    return [];
  }
});

/**
 * 聊天 API 測試
 * @param {string} baseUrl - API 基礎 URL
 * @param {Object} session - 當前用戶的會話（user1）
 * @returns {Object} { dmRoomId, chatWithUserId }
 */
export default function (baseUrl, session) {
  if (!session) return null;

  const groupStartTime = Date.now();
  logGroupStart('Chat Management APIs');

  let dmRoomId;
  let chatWithUserId;
  let skipTests = false;

  group('API - Chat', function () {
    // 步驟 1: 獲取第二個測試用戶的 ID（通過測試 API）
    group('Get Second User ID', function () {
      // 從 testUsers 中選擇一個不同的用戶
      let user2;
      if (testUsers.length >= 2) {
        // 找一個不是當前用戶的用戶
        user2 = testUsers.find(u => u.email !== session.user.email) || testUsers[1];
      } else {
        logError('測試用戶數量不足，需要至少 2 個用戶');
        skipTests = true;
        return;
      }
      
      logInfo(`使用第二個測試用戶: ${user2.username} (${user2.email})`);
      
      // 確保 user2 已註冊（不需要登入）
      registerUser(baseUrl, user2);
      
      // 🎯 使用新的測試 API：GET /test/user?username=xxx
      // 不需要認證，直接通過 username 取得用戶 ID
      // 使用完整 URL 避免路徑問題
      const testApiUrl = `${baseUrl}/test/user?username=${user2.username}`;
      const userRes = http.get(testApiUrl, {
        headers: {
          ...TEST_CONFIG.DEFAULT_HEADERS,
        }
      });
      let userBody;
      try {
        userBody = userRes.json();
      } catch (e) {
        userBody = null;
      }
      
      logHttpResponse('GET /test/user', userRes, { expectedStatus: 200 });
      
      if (check(userRes, {
        'Get User by Username: status is 200': (r) => r.status === 200,
        'Get User by Username: has user data': () => userBody && userBody.data && userBody.data.id !== undefined
      })) {
        chatWithUserId = userBody.data.id;
        logInfo(`✅ 取得第二個用戶 ID: ${chatWithUserId} (${user2.username})`);
      } else {
        logError('❌ 無法取得第二個用戶 ID');
        logError(`回應: ${userRes.body}`);
        skipTests = true;
      }
    });

    // 如果沒有成功取得 user2 ID，跳過後續測試
    if (skipTests || !chatWithUserId) {
      logError('⚠️  無法取得聊天對象用戶 ID，跳過私聊測試');
    } else {
      // 步驟 2: 取得 DM 房間列表
      group('Get DM Rooms', function () {
        logInfo('取得私聊房間列表');
        const headers = {
          ...session.headers,
          ...TEST_CONFIG.DEFAULT_HEADERS,
        };
        const res = http.get(`${baseUrl}/dm_rooms`, { headers });
        
        logHttpResponse('GET /dm_rooms', res, { expectedStatus: 200 });
        
        if (check(res, { 
          'Get DM Rooms: status is 200': (r) => r.status === 200,
          'Get DM Rooms: response has success status': (r) => r.json('status') === 'success'
        })) {
          // 嘗試提取第一個 DM 房間的 ID
          const dmRooms = res.json('data');
          if (dmRooms && dmRooms.length > 0 && dmRooms[0].room_id) {
            dmRoomId = dmRooms[0].room_id;
            logInfo(`✅ 找到現有 DM 房間，ID: ${dmRoomId}`);
          } else {
            logInfo('ℹ️  當前沒有 DM 房間');
          }
        }
      });

      // 步驟 3: 創建與 user2 的私聊房間
      group('Create DM Room', function () {
        const user2Username = testUsers.find(u => u.email !== session.user.email)?.username || 'testuser1';
        logInfo(`創建私聊房間與用戶: ${chatWithUserId} (${user2Username})`);
        
        const url = `${baseUrl}/dm_rooms`;
        const payload = JSON.stringify({
          chat_with_user_id: chatWithUserId // 使用真實的 user2 ID
        });
        const headers = {
          'Content-Type': 'application/json',
          ...session.headers,
          ...applyCsrf(url, {}),
          ...TEST_CONFIG.DEFAULT_HEADERS,
        };
        const res = http.post(url, payload, { headers });
        
        logHttpResponse('POST /dm_rooms', res, { expectedStatus: [200, 400] });
        
        if (check(res, { 
          'Create DM Room: status is 200 or 400': (r) => r.status === 200 || r.status === 400,
          'Create DM Room: response has status field': (r) => r.json('status') !== undefined
        })) {
          if (res.status === 200 && res.json('data.room_id')) {
            dmRoomId = res.json('data.room_id');
            logInfo(`✅ 私聊房間創建成功，ID: ${dmRoomId}`);
          } else if (res.status === 400) {
            // 可能房間已存在
            const errorMessage = res.json('message') || '';
            if (errorMessage.includes('已存在') || errorMessage.includes('exist')) {
              logInfo('ℹ️  私聊房間已存在');
              // 重新獲取房間列表以取得房間 ID
              const listRes = http.get(`${baseUrl}/dm_rooms`, { headers: session.headers });
              if (listRes.status === 200) {
                const dmRooms = listRes.json('data');
                if (dmRooms && dmRooms.length > 0) {
                  dmRoomId = dmRooms[0].room_id;
                  logInfo(`✅ 從列表中取得 DM 房間 ID: ${dmRoomId}`);
                }
              }
            } else {
              logInfo(`ℹ️  創建失敗: ${errorMessage}`);
            }
          }
        }
      });

      // 步驟 4: 更新 DM 房間狀態（如果有房間 ID）
      if (dmRoomId) {
        group('Update DM Room Status', function () {
          logInfo(`更新私聊房間狀態: ${dmRoomId} (設為可見)`);
          
          const url = `${baseUrl}/dm_rooms`;
          const payload = JSON.stringify({
            room_id: dmRoomId,
            is_hidden: false
          });
          const headers = {
            'Content-Type': 'application/json',
            ...session.headers,
            ...applyCsrf(url, {}),
            ...TEST_CONFIG.DEFAULT_HEADERS,
          };
          const res = http.put(url, payload, { headers });
          
          logHttpResponse('PUT /dm_rooms', res, { expectedStatus: [200, 400, 500] });
          
          if (check(res, { 
            'Update DM Room Status: status is 200, 400 or 500': (r) => r.status === 200 || r.status === 400 || r.status === 500,
            'Update DM Room Status: response has status field': (r) => r.json('status') !== undefined
          })) {
            if (res.status === 200) {
              logInfo('✅ 私聊房間狀態更新成功');
            } else if (res.status === 400) {
              logInfo('ℹ️  更新失敗 (可能房間不存在或參數錯誤)');
            } else if (res.status === 500) {
              logInfo('ℹ️  伺服器錯誤');
            }
          }
        });

        // 步驟 5: 獲取 DM 房間訊息
        group('Get DM Messages', function () {
          logInfo(`取得私聊房間訊息: ${dmRoomId} (限制10條)`);
          
          const headers = {
            ...session.headers,
            ...TEST_CONFIG.DEFAULT_HEADERS,
          };
          const res = http.get(`${baseUrl}/dm_rooms/${dmRoomId}/messages?limit=10`, { headers });
          
          logHttpResponse('GET /dm_rooms/{id}/messages', res, { expectedStatus: [200, 404, 500] });
          
          if (check(res, { 
            'Get DM Messages: status is 200, 404 or 500': (r) => r.status === 200 || r.status === 404 || r.status === 500,
            'Get DM Messages: response has status field': (r) => r.json('status') !== undefined
          })) {
            if (res.status === 200) {
              const messages = res.json('data') || [];
              logInfo(`✅ 私聊訊息取得成功，共 ${messages.length} 條訊息`);
            } else if (res.status === 404) {
              logInfo('ℹ️  私聊房間不存在或無權限存取');
            } else if (res.status === 500) {
              logInfo('ℹ️  伺服器錯誤');
            }
          }
        });
      } else {
        logInfo('⚠️  沒有可用的 DM 房間 ID，跳過房間操作測試');
      }
    }
  });
  
  logGroupEnd('Chat Management APIs', groupStartTime);
  
  // 返回 DM 房間 ID 和聊天對象 ID
  return {
    dmRoomId,
    chatWithUserId
  };
}
