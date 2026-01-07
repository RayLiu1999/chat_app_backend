/**
 * 認證輔助函數
 */
import http from "k6/http";
import { check } from "k6";
import { extractCSRFToken } from "./csrf.js";
import { randomString } from "./utils.js";
import { logHttpResponse, logInfo, logError } from "./logger.js";
import { SharedArray } from "k6/data";

const testUsers = new SharedArray("testUsers", function () {
  try {
    // 相對於本檔案位置: loadtest/scripts/common/auth.js → loadtest/data/users.json
    return JSON.parse(open("../../data/users.json"));
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
      "Content-Type": "application/json",
    },
    // 使用 expectedStatuses 將 200 和 400 標記為成功（避免 http_req_failed 計算錯誤）
    responseType: "text",
  };

  logInfo(`註冊用戶: ${user.email} (${user.username})`);

  // 使用 http.expectedStatuses 來標記 200 和 400 為預期狀態碼
  const res = http.post(url, payload, {
    ...params,
    responseCallback: http.expectedStatuses(200, 400),
  });

  // 記錄 HTTP 回應
  logHttpResponse("POST /register", res, { expectedStatus: [200, 400] });

  let body;
  try {
    body = res.json();
  } catch (e) {
    body = null;
  }
  const isSuccess = res.status === 200 && body && body.status === "success";
  // 檢查是否為預期的錯誤（用戶已存在），同時驗證 HTTP 狀態碼和錯誤代碼
  const isUserExists =
    res.status === 400 &&
    body &&
    body.status === "error" &&
    (body.code === "USERNAME_EXISTS" || body.code === "EMAIL_EXISTS");

  check(res, {
    "register: status is 200 or 400": (r) =>
      r.status === 200 || r.status === 400,
    "register: success or user already exists": () => isSuccess || isUserExists,
  });

  if (isSuccess) {
    logInfo("✅ 用戶註冊成功");
  } else if (isUserExists) {
    logInfo("ℹ️  用戶已存在，跳過註冊");
  } else {
    logError("❌ 用戶註冊失敗", `Status: ${res.message}`);
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
      "Content-Type": "application/json",
    },
  };

  logInfo(`嘗試登入: ${credentials.email}`);
  const res = http.post(url, payload, params);

  // 記錄 HTTP 回應
  logHttpResponse("POST /login", res, { expectedStatus: [200, 400, 401] });

  let body;
  try {
    body = res.json();
  } catch (e) {
    body = null;
  }
  const isSuccess = res.status === 200 && body && body.status === "success";
  // 登入失敗通常是 400 或 401，且 status 為 error
  const isInvalidCredentials =
    (res.status === 400 || res.status === 401) &&
    body &&
    body.status === "error";

  check(res, {
    "login: status is 200, 400 or 401": (r) =>
      [200, 400, 401].includes(r.status),
    "login: response has valid JSON": (r) => body !== null,
    "login: success or invalid credentials": (r) =>
      isSuccess || isInvalidCredentials,
  });

  if (isSuccess) {
    const token = res.json("data.access_token");

    // 提取 CSRF Token（從 Set-Cookie header）
    const csrfToken = extractCSRFToken(res, url);

    if (!csrfToken) {
      logError("❌ 無法提取 CSRF Token");
    }

    logInfo("🔐 登入成功，取得 Access Token 和 CSRF Token");

    return {
      token,
      csrfToken,
    };
  } else if (isInvalidCredentials) {
    logInfo("ℹ️  登入失敗，用戶名稱或密碼錯誤");
    return null;
  } else {
    logError(
      "❌ 登入失敗",
      `Status: ${res.status}, Response: ${
        res.body ? res.body.substring(0, 100) : "empty"
      }`
    );
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
  let user =
    testUsers.length > 0
      ? testUsers[__VU % testUsers.length]
      : {
          username: `user_${randomString(6)}`,
          email: `user_${randomString(6)}@example.com`,
          password: "Password123!",
          nickname: `User ${randomString(6)}`,
        };

  // 先嘗試登入
  let loginResult = login(baseUrl, {
    email: user.email,
    password: user.password,
  });

  if (!loginResult) {
    // 嘗試註冊後再登入
    registerUser(baseUrl, user);
    loginResult = login(baseUrl, {
      email: user.email,
      password: user.password,
    });
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
