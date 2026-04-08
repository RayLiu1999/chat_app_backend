/**
 * 即時日誌工具 - 用於所有測試腳本的統一日誌輸出
 */
import { Counter, Rate } from "k6/metrics";

// 全域計數器
export const apiRequestCounter = new Counter("api_requests_logged");
export const apiSuccessCounter = new Counter("api_success_logged");
export const apiErrorCounter = new Counter("api_errors_logged");

// 檢查是否啟用詳細模式
const VERBOSE_MODE = __ENV.VERBOSE === "1";

/**
 * 記錄 HTTP 請求結果
 * @param {string} apiName - API 名稱 (例如: "POST /auth/login")
 * @param {Object} response - k6 HTTP 回應物件
 * @param {Object} options - 額外選項
 */
export function logHttpResponse(apiName, response, options = {}) {
  const timestamp = new Date().toISOString().slice(11, 23); // 只顯示時間部分
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const iter = typeof __ITER !== "undefined" ? __ITER : 0;

  // 增加計數器
  apiRequestCounter.add(1);

  // 基本資訊
  const status = response.status;
  const duration = response.timings.duration.toFixed(2);
  const size = response.body ? (response.body.length / 1024).toFixed(2) : "0";

  // 判斷成功或失敗
  const isSuccess = status >= 200 && status < 400;
  const statusIcon = isSuccess ? "✅" : "❌";

  if (isSuccess) {
    apiSuccessCounter.add(1);
  } else {
    apiErrorCounter.add(1);
  }

  // 構建日誌訊息
  let logMessage = `[${timestamp}] [VU:${vu}] [Iter:${iter}] ${statusIcon} ${apiName}`;
  logMessage += ` | Status: ${status}`;
  logMessage += ` | Duration: ${duration}ms`;
  logMessage += ` | Size: ${size}KB`;

  // 詳細模式：顯示請求/回應內容∏
  if (VERBOSE_MODE) {
    try {
      const responseData = response.json();
      logMessage += ` | Response: ${JSON.stringify(responseData).substring(
        0,
        200,
      )}`;
      if (JSON.stringify(responseData).length > 200) {
        logMessage += "...";
      }
    } catch (e) {
      // 如果回應不是 JSON，顯示前 100 個字元
      const bodyPreview = response.body ? response.body.substring(0, 100) : "";
      if (bodyPreview) {
        logMessage += ` | Body: ${bodyPreview}`;
        if (response.body.length > 100) {
          logMessage += "...";
        }
      }
    }
  }

  // 如果有錯誤，加上錯誤詳情
  if (!isSuccess) {
    logMessage += ` | URL: ${response.url}`;
    if (options.expectedStatus) {
      logMessage += ` | Expected: ${options.expectedStatus}`;
    }
  }

  console.log(logMessage);

  return {
    success: isSuccess,
    status: status,
    duration: parseFloat(duration),
    size: parseFloat(size),
  };
}

/**
 * 記錄 WebSocket 事件
 * @param {string} event - 事件名稱
 * @param {string} details - 詳細資訊
 * @param {Object} data - 相關資料
 */
export function logWebSocketEvent(event, details, data = null) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const iter = typeof __ITER !== "undefined" ? __ITER : 0;

  let logMessage = `[${timestamp}] [VU:${vu}] [Iter:${iter}] 🔌 WebSocket ${event}`;

  if (details) {
    logMessage += ` | ${details}`;
  }

  if (data && VERBOSE_MODE) {
    logMessage += ` | Data: ${JSON.stringify(data).substring(0, 150)}`;
    if (JSON.stringify(data).length > 150) {
      logMessage += "...";
    }
  }

  console.log(logMessage);
}

/**
 * 記錄一般資訊
 * @param {string} message - 訊息內容
 * @param {Object} data - 相關資料
 */
export function logInfo(message, data = null) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const iter = typeof __ITER !== "undefined" ? __ITER : 0;

  let logMessage = `[${timestamp}] [VU:${vu}] [Iter:${iter}] ℹ️  ${message}`;

  if (data && VERBOSE_MODE) {
    logMessage += ` | ${JSON.stringify(data)}`;
  }

  console.log(logMessage);
}

/**
 * 記錄成功資訊
 * @param {string} message - 成功訊息
 * @param {Object} data - 相關資料
 * @param {number} duration - 耗時 (毫秒)
 */
export function logSuccess(message, data = null, duration = null) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const iter = typeof __ITER !== "undefined" ? __ITER : 0;

  let logMessage = `[${timestamp}] [VU:${vu}] [Iter:${iter}] ✅ SUCCESS: ${message}`;

  if (duration !== null) {
    logMessage += ` | Duration: ${duration}ms`;
  }

  if (data && VERBOSE_MODE) {
    logMessage += ` | ${JSON.stringify(data)}`;
  }

  console.log(logMessage);
}

/**
 * 記錄錯誤資訊
 * @param {string} message - 錯誤訊息
 * @param {Object} error - 錯誤物件
 */
export function logError(message, error = null) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const iter = typeof __ITER !== "undefined" ? __ITER : 0;

  let logMessage = `[${timestamp}] [VU:${vu}] [Iter:${iter}] 🚨 ERROR: ${message}`;

  if (error) {
    if (typeof error === "string") {
      logMessage += ` | ${error}`;
    } else if (error.message) {
      logMessage += ` | ${error.message}`;
    } else {
      logMessage += ` | ${JSON.stringify(error)}`;
    }
  }

  console.error(logMessage);
}

/**
 * 記錄測試群組開始
 * @param {string} groupName - 群組名稱
 */
export function logGroupStart(groupName) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;

  console.log(`[${timestamp}] [VU:${vu}] 📂 開始測試群組: ${groupName}`);
  console.log("─".repeat(50));
}

/**
 * 記錄測試群組結束
 * @param {string} groupName - 群組名稱
 * @param {number} startTime - 開始時間
 */
export function logGroupEnd(groupName, startTime) {
  const timestamp = new Date().toISOString().slice(11, 23);
  const vu = typeof __VU !== "undefined" ? __VU : 0;
  const duration = Date.now() - startTime;

  console.log("─".repeat(50));
  console.log(
    `[${timestamp}] [VU:${vu}] 📂 完成測試群組: ${groupName} | 耗時: ${duration}ms`,
  );
  console.log("");
}
