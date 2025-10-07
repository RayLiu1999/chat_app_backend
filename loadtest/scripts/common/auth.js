/**
 * 認證輔助函數
 */
import http from 'k6/http';
import { check } from 'k6';
import { extractCSRFToken, applyCsrf } from './csrf.js';
import { randomString } from './utils.js';
import { logHttpResponse, logInfo, logError } from './logger.js';
import { SharedArray } from 'k6/data';

const testUsers = new SharedArray('testUsers', function () {
  try {
    // 相對於本檔案位置: loadtest/scripts/common/auth.js → loadtest/data/users.json
    return JSON.parse(open('../../data/users.json'));
  } catch (e) {
    return [];
  }
});

/**
 * 註冊一個新用戶
 * @param {string} baseUrl - API 基礎 URL
 * @param {Object} user - 用戶物件 { username, email, password, nickname }
 * @returns {boolean} 是否註冊成功或已存在
 */
export function registerUser(baseUrl, user) {
  const url = `${baseUrl}/register`;
  const payload = JSON.stringify({
    username: user.username,
    email: user.email,
    password: user.password,
    nickname: user.nickname || user.username, // 如果沒有 nickname 就使用 username
  });
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  logInfo(`註冊用戶: ${user.email} (${user.username})`);
  const res = http.post(url, payload, params);
  
  // 記錄 HTTP 回應
  logHttpResponse('POST /register', res, { expectedStatus: [200, 400] });

  const isSuccess = res.json('status') === 'success';
  const isUserExists = res.json('status') === 'error' && (res.json('code') == 'USERNAME_EXISTS');

  check(res, {
      'register: success': () => isSuccess,
      'register: user already exists': () => isUserExists,
  });

  if (isSuccess) {
    logInfo('✅ 用戶註冊成功');
  }
  else if (isUserExists) {
    logInfo('ℹ️  用戶已存在，跳過註冊');
  }
  else {
    logError('❌ 用戶註冊失敗', `Status: ${res.message}`);
  }

  return { isSuccess, isUserExists };
}

/**
 * 使用者登入並取得 token 和 CSRF token
 * @param {string} baseUrl - API 基礎 URL
 * @param {Object} credentials - 登入憑證 { email, password }
 * @returns {Object|null} { token, csrfToken } 或 null
 */
export function login(baseUrl, credentials) {
  const url = `${baseUrl}/login`;
  const payload = JSON.stringify(credentials);
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  logInfo(`嘗試登入: ${credentials.email}`);
  const res = http.post(url, payload, params);
  
  // 記錄 HTTP 回應
  logHttpResponse('POST /login', res, { expectedStatus: 200 }); 

  const isSuccess = res.json('status') === 'success';
  const isInvalidCredentials = res.json('status') === 'error';

  check(res, {
    'login: success': (r) => isSuccess,
    'login: invalid credentials': (r) => isInvalidCredentials,
  });
  
  if (isSuccess) {
    const token = res.json('data.access_token');
    
    // 提取 CSRF Token（從 Set-Cookie header）
    const csrfToken = extractCSRFToken(res, url);
    
    if (!csrfToken) {
      logError('❌ 無法提取 CSRF Token');
    }
    
    logInfo('🔐 登入成功，取得 Access Token 和 CSRF Token');
    
    return {
      token,
      csrfToken,
    };
  } else if (isInvalidCredentials) {
    logInfo('ℹ️  登入失敗，用戶名稱或密碼錯誤');
    return null;
  } else {
    logError('❌ 登入失敗', `Status: ${res.status}, Response: ${res.body.substring(0, 100)}`);
    return null;
  }
}

/**
 * 獲取一個已認證的用戶 session (token, csrfToken 和 headers)
 * @param {string} baseUrl - API 基礎 URL
 * @returns {Object|null} 包含 token, csrfToken 和 headers 的 session 物件
 */
export function getAuthenticatedSession(baseUrl) {
  // 取測試用戶或產生臨時用戶
  let user = testUsers.length > 0
    ? testUsers[__VU % testUsers.length]
    : { 
        username: `user_${randomString(6)}`, 
        email: `user_${randomString(6)}@example.com`, 
        password: 'Password123!',
        nickname: `User ${randomString(6)}`
      };

  // 先嘗試登入
  let loginResult = login(baseUrl, { email: user.email, password: user.password });
  
  if (!loginResult) {
    // 嘗試註冊後再登入
    registerUser(baseUrl, user);
    loginResult = login(baseUrl, { email: user.email, password: user.password });
  }

  if (loginResult && loginResult.token) {
    return {
      token: loginResult.token,
      csrfToken: loginResult.csrfToken,
      headers: {
        Authorization: `Bearer ${loginResult.token}`,
      },
      user: user, // 返回用戶資訊以便其他測試使用
    };
  }
  
  return null;
}
