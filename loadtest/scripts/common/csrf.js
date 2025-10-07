/**
 * CSRF Token 處理
 * 
 * 後端機制：
 * 1. 登入/刷新 Token 時，後端會將 CSRF Token 寫入 cookie（csrf_token）
 * 2. 前端在非 GET 請求時，需要從 cookie 讀取並在 Header 中帶上 X-CSRF-TOKEN
 * 3. 後端驗證 Header 中的 token 和 Cookie 中的 token 是否一致
 */
import http from 'k6/http';
import { logInfo, logError } from './logger.js';

/**
 * 從登入回應中提取 CSRF Token
 * 後端會在登入時設置 csrf_token cookie
 * 
 * @param {Object} response - HTTP 回應物件
 * @param {string} url - 請求的 URL（用於設置 cookie）
 * @returns {string|null} CSRF Token 或 null
 */
export function extractCSRFToken(response, url) {
  try {
    // 從回應的 Set-Cookie header 中提取 csrf_token
    const setCookieHeaders = response.headers['Set-Cookie'];
    
    if (!setCookieHeaders) {
      logError('未找到 Set-Cookie header');
      return null;
    }
    
    // Set-Cookie 可能是字串或陣列
    const cookies = Array.isArray(setCookieHeaders) ? setCookieHeaders : [setCookieHeaders];
    
    // 尋找 csrf_token cookie
    for (const cookie of cookies) {
      if (cookie.includes('csrf_token=')) {
        // 提取 token 值（格式: csrf_token=xxx; Path=/; ...)
        const match = cookie.match(/csrf_token=([^;]+)/);
        if (match && match[1]) {
          const token = match[1];
          logInfo(`✅ CSRF Token 提取成功: ${token.substring(0, 20)}...`);
          
          // 將 token 儲存到 k6 的 cookie jar
          const jar = http.cookieJar();
          jar.set(url, 'csrf_token', token);
          
          return token;
        }
      }
    }
    
    logError('未找到 csrf_token cookie');
    return null;
  } catch (e) {
    logError('提取 CSRF Token 失敗', e.message);
    return null;
  }
}

/**
 * 為請求應用 CSRF Token
 * 
 * @param {string} url - 目標 URL
 * @param {Object} headers - 原始標頭
 * @param {string} csrfToken - CSRF Token（可選，如果不提供則從 cookie 讀取）
 * @returns {Object} 包含 CSRF 標頭和 Cookie 的新標頭物件
 */
export function applyCsrf(url, headers = {}, csrfToken = null) {
  try {
    // 如果沒有提供 token，嘗試從 cookie jar 讀取
    if (!csrfToken) {
      const jar = http.cookieJar();
      const cookies = jar.cookiesForURL(url);
      
      // 從 cookies 物件中找到 csrf_token
      for (const [name, values] of Object.entries(cookies)) {
        if (name === 'csrf_token' && values.length > 0) {
          csrfToken = values[0];
          break;
        }
      }
    }
    
    if (!csrfToken) {
      logError('無法取得 CSRF Token，請求可能會失敗');
      return headers;
    }
    
    // 返回包含 CSRF Token 的 headers
    // 注意：k6 會自動從 cookie jar 發送 cookie，所以不需要手動設置 Cookie header
    return {
      ...headers,
      'X-CSRF-TOKEN': csrfToken,
    };
  } catch (e) {
    logError('應用 CSRF Token 失敗', e.message);
    return headers;
  }
}

/**
 * 取得當前的 CSRF Token
 * 
 * @param {string} url - 目標 URL
 * @returns {string|null} CSRF Token 或 null
 */
export function getCurrentCSRFToken(url) {
  try {
    const jar = http.cookieJar();
    const cookies = jar.cookiesForURL(url);
    
    for (const [name, values] of Object.entries(cookies)) {
      if (name === 'csrf_token' && values.length > 0) {
        return values[0];
      }
    }
    
    return null;
  } catch (e) {
    logError('取得 CSRF Token 失敗', e.message);
    return null;
  }
}
