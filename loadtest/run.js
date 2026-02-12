/**
 * k6 測試主執行文件
 *
 * 使用方法:
 * k6 run run.js --env SCENARIO=smoke
 * k6 run run.js --env SCENARIO=monolith_capacity
 *
 * 參數:
 * --env SCENARIO: 要執行的測試場景 (smoke, monolith_capacity)
 * --env BASE_URL: 覆蓋 config.js 中的 API URL (預設: http://localhost:80)
 * --env WS_URL: 覆蓋 WebSocket URL (預設: ws://localhost:80/ws)
 * --env VERBOSE: 啟用詳細日誌模式 (1 為啟用)
 * --out json=test_results/load_tests/api/output.json: 將結果輸出到檔案
 */
import { SharedArray } from 'k6/data';
import http from "k6/http";
import { Counter, Rate } from 'k6/metrics';
import * as config from './config.js';
import smokeTest from './scenarios/smoke.js';
import monolithCapacityTest from './scenarios/monolith_capacity.js';
import { getAuthenticatedSession } from './scripts/common/auth.js';
import { logInfo, logSuccess, logError } from './scripts/common/logger.js';

// 自定義metrics用於即時監控
export const apiRequestCount = new Counter('api_requests_total');
export const apiSuccessRate = new Rate('api_success_rate');
export const wsConnectionCount = new Counter('ws_connections_total');

// 設定詳細日誌模式
const VERBOSE_MODE = __ENV.VERBOSE === '1';

// 選擇測試場景
const scenarioName = __ENV.SCENARIO || 'smoke';
const scenarios = {
  smoke: smokeTest,
  monolith_capacity: monolithCapacityTest,
};

if (!scenarios[scenarioName]) {
  throw new Error(`無效的場景名稱: ${scenarioName}. 可用選項: ${Object.keys(scenarios).join(', ')}`);
}

// 設置 k6 選項（使用 stages 配置對應 ramping-vus 預設執行器）
export const options = {
  scenarios: {
    [scenarioName]: {
      executor: 'ramping-vus',   // stages 本質上就是 ramping-vus
      stages: config.TEST_CONFIG.SCENARIOS[scenarioName],
      gracefulStop: '5s',        // 🔑 關鍵設定，避免多等 30 秒
      gracefulRampDown: '5s'
    },
  },
  thresholds: config.TEST_CONFIG.THRESHOLDS,
  // 確保有足夠的迭代時間
  maxRedirects: 10,
  // 設定 UserAgent
  userAgent: 'k6-load-test/1.0',
  // 啟用即時摘要輸出
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(95)', 'p(99)', 'count'],
  // 每 10 秒輸出一次狀態
  summaryTimeUnit: 'ms',
};

// 設置迭代初始化（所有 VU 開始前執行一次）
export function setup() {
  console.log(`🚀 開始執行 ${scenarioName} 測試場景`);
  console.log(`📍 API 基礎 URL: ${config.TEST_CONFIG.BASE_URL}`);
  console.log(`🔌 WebSocket URL: ${config.TEST_CONFIG.WS_URL}`);
  console.log(`📝 詳細日誌模式: ${VERBOSE_MODE ? '啟用' : '停用'}`);
  console.log('=' .repeat(60));
  
  // 在這裡執行一次性的身份驗證，避免 VUs 重複執行 Bcrypt
  let session = null;
  try {
     session = getAuthenticatedSession(`${config.TEST_CONFIG.BASE_URL}${config.TEST_CONFIG.API_PREFIX}`);
      if (session && session.token) {
       console.log(`🔑 身份驗證成功，Token: ${session.token.substring(0, 10)}...`);
      } else {
       console.warn(`⚠️ 身份驗證未返回有效 Session`);
      }
  } catch (e) {
     console.error(`❌ 身份驗證失敗: ${e.message}`);
  }

  return {
    startTime: Date.now(),
    scenario: scenarioName,
    verbose: VERBOSE_MODE,
    session: session,
    config: config.TEST_CONFIG
  };
}

// 主執行函數
export default function (data) {
  const iterationStart = Date.now();
  
  // 同步 CSRF Cookie 到當前 VU 的 Cookie Jar (解決 setup() 資料不會自動同步 Cookie 的問題)
  if (data.session && data.session.csrfToken) {
    const jar = http.cookieJar();
    jar.set(data.config.BASE_URL, "csrf_token", data.session.csrfToken);
  }

  logInfo(`開始執行迭代 - 場景: ${data?.scenario || scenarioName}`);
  
  try {
    // 傳入 config 和 session，保持腳本相容性
    // 注意: session 可能為 null，由各場景自行處理
    scenarios[scenarioName](data.config, data.session);
    
    const duration = Date.now() - iterationStart;
    logSuccess(`迭代完成`, null, duration);
    
  } catch (error) {
    logError(`迭代執行失敗`, error.message);
    throw error;
  }
}

// 測試清理（所有 VU 完成後執行一次）
export function teardown(data) {
  const endTime = Date.now();
  const totalDuration = endTime - data.startTime;
  
  console.log('=' .repeat(60));
  console.log(`🏁 測試場景 ${data.scenario} 執行完畢`);
  console.log(`⏱️  總執行時間: ${totalDuration}ms (${(totalDuration / 1000).toFixed(2)}秒)`);
  console.log('=' .repeat(60));
}

// 測試結束時輸出摘要
export function handleSummary(data) {
  const now = new Date().toISOString().replace(/[:.]/g, '').replace('T', '_').slice(0, 15);
  const scenario = __ENV.SCENARIO || 'smoke';
  const outputDir = config.TEST_CONFIG.RESULTS_DIR;

  const getMetricValue = (metricName, type = 'value') => {
    if (data.metrics[metricName] && data.metrics[metricName].values) {
      return data.metrics[metricName].values[type] || 0;
    }
    return 0;
  };

  const passes = getMetricValue('checks', 'passes');
  const fails = getMetricValue('checks', 'fails');
  const total = passes + fails;
  const totalRequests = getMetricValue('http_reqs', 'count');
  const successRequests = getMetricValue('http_req_failed', 'fails');
  const failureRate = getMetricValue('http_req_failed', 'rate');
  const failedRequests = totalRequests - successRequests;
  const avgDuration = getMetricValue('http_req_duration', 'avg');
  const p95Duration = getMetricValue('http_req_duration', 'p(95)');

  // 建立一個基本的 markdown 報告
  let report = `# k6 測試報告: ${scenario}\n\n`;
  report += `**測試時間:** ${new Date().toLocaleString()}\n\n`;
  report += '## 測試結果摘要\n\n';
  report += `* **check 成功/總數:** ${passes}/${total}\n`;
  report += `* **check 成功率:** ${(total > 0 ? (passes / total) * 100 : 0).toFixed(2)}%\n`;
  report += `* **HTTP 總請求數:** ${totalRequests}\n`;
  report += `* **HTTP 成功請求數:** ${successRequests}\n`;
  report += `* **HTTP 失敗請求數:** ${failedRequests}\n`;
  report += `* **HTTP 請求失敗率:** ${(failureRate * 100).toFixed(2)}%\n`;
  report += `* **HTTP 平均響應時間:** ${avgDuration.toFixed(2)}ms\n`;
  report += `* **HTTP p(95) 響應時間:** ${p95Duration.toFixed(2)}ms\n`;
  
  if (data.metrics.ws_connecting && data.metrics.ws_connecting.values) {
    const wsP95 = data.metrics.ws_connecting.values['p(95)'] || 0;
    report += `* **WebSocket p(95) 連接時間:** ${wsP95.toFixed(2)}ms\n`;
  }

  return {
    stdout: report,
    [`${outputDir}/${now}_${scenario}_summary.md`]: report,
    [`${outputDir}/${now}_${scenario}_results.json`]: JSON.stringify(data, null, 2),
  };
}
