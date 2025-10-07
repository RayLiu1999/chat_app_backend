/**
 * 輕量負載測試場景
 * 目的：模擬少量用戶同時在線，進行常規操作。
 */
import apiAuth from '../scripts/api/auth.js';
import apiServers from '../scripts/api/servers.js';
import apiFriends from '../scripts/api/friends.js';
import apiChat from '../scripts/api/chat.js';
import apiUpload from '../scripts/api/upload.js';
import { getAuthenticatedSession } from '../scripts/common/auth.js';
import { randomSleep } from '../scripts/common/utils.js';

export default function (config) {
  const baseUrl = config.BASE_URL;
  
  const session = getAuthenticatedSession(baseUrl);
  if (session) {
    // 隨機執行 API 操作，根據 API.md 的端點進行測試
    const actions = [
      () => apiServers(baseUrl, session),
      () => apiFriends(baseUrl, session),
      () => apiChat(baseUrl, session),
      () => apiUpload(baseUrl, session),
    ];
    
    const action = actions[Math.floor(Math.random() * actions.length)];
    action();
    
    randomSleep(2, 5);
  } else {
    // 如果登入失敗，只做註冊和登入測試
    apiAuth(baseUrl);
  }
}
