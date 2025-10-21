/**
 * k6 測試主執行文件
 *
 * 使用方法:
 * k6 run run.js --env SCENARIO=smoke
 * k6 run run.js --env SCENARIO=light
 * k6 run run.js --env SCENARIO=medium
 * k6 run run.js --env SCENARIO=heavy
 * k6 run run.js --env SCENARIO=websocket_stress
 * k6 run run.js --env SCENARIO=websocket_spike
 * k6 run run.js --env SCENARIO=websocket_soak
 * k6 run run.js --env SCENARIO=websocket_stress_ladder
 *
 * 參數:
 * --env SCENARIO: 要執行的測試場景 (smoke, light, medium, heavy)
 * --env BASE_URL: 覆蓋 config.js 中的 API URL (預設: http://localhost:8080)
 * --env WS_URL: 覆蓋 WebSocket URL (預設: ws://localhost:8080/ws)
 * --env VERBOSE: 啟用詳細日誌模式 (1 為啟用)
 * --out json=test_results/load_tests/api/output.json: 將結果輸出到檔案
 */
import { SharedArray } from 'k6/data';
import { Counter, Rate } from 'k6/metrics';
import * as config from './config.js';
import smokeTest from './scenarios/smoke.js';
import lightLoadTest from './scenarios/light.js';
import mediumLoadTest from './scenarios/medium.js';
import heavyLoadTest from './scenarios/heavy.js';
import websocketStressTest from './scenarios/websocket_stress.js';
import websocketSpikeTest from './scenarios/websocket_spike.js';
import websocketSoakTest from './scenarios/websocket_soak.js';
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
  light: lightLoadTest,
  medium: mediumLoadTest,
  heavy: heavyLoadTest,
  websocket_stress: websocketStressTest,
  websocket_spike: websocketSpikeTest,
  websocket_soak: websocketSoakTest,
  websocket_stress_ladder: websocketStressTest, // 使用相同的測試函數，但不同的 stages
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

// 設置迭代初始化（每個 VU 開始時執行一次）
export function setup() {
  console.log(`🚀 開始執行 ${scenarioName} 測試場景`);
  console.log(`📍 API 基礎 URL: ${config.TEST_CONFIG.BASE_URL}`);
  console.log(`🔌 WebSocket URL: ${config.TEST_CONFIG.WS_URL}`);
  console.log(`📝 詳細日誌模式: ${VERBOSE_MODE ? '啟用' : '停用'}`);
  console.log('=' .repeat(60));
  
  return {
    startTime: Date.now(),
    scenario: scenarioName,
    verbose: VERBOSE_MODE
  };
}

// 主執行函數
export default function (data) {
  const iterationStart = Date.now();
  
  logInfo(`開始執行迭代 - 場景: ${data?.scenario || scenarioName}`);
  
  try {
    scenarios[scenarioName](config.TEST_CONFIG);
    
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
  const outputDir = 'test_results/load_tests';  // 使用相對路徑

  const passes = data.metrics.checks.values.passes || 0;
  const fails = data.metrics.checks.values.fails || 0;
  const total = passes + fails;
  const totalRequests = data.metrics.http_reqs.values.count || 0;
  const failedRequests = data.metrics.http_req_failed.values.fails || 0;

  // 建立一個基本的 markdown 報告
  let report = `# k6 測試報告: ${scenario}\n\n`;
  report += `**測試時間:** ${new Date().toLocaleString()}\n\n`;
  report += '## 測試結果摘要\n\n';
  report += `* **check 成功/總數:** ${passes}/${total}\n`;
  report += `* **check 成功率:** ${((passes / total) * 100).toFixed(2)}%\n`;
  report += `* **HTTP 總請求數:** ${totalRequests}\n`;
  report += `* **HTTP 失敗請求數:** ${failedRequests}\n`;
  report += `* **HTTP 請求失敗率:** ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  report += `* **HTTP 平均響應時間:** ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  report += `* **HTTP p(95) 響應時間:** ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  if (data.metrics.ws_connecting) {
    report += `* **WebSocket p(95) 連接時間:** ${data.metrics.ws_connecting.values['p(95)'].toFixed(2)}ms\n`;
  }

  return {
    stdout: report,
    [`${outputDir}/${now}_${scenario}_summary.md`]: report,
    [`${outputDir}/${now}_${scenario}_results.json`]: JSON.stringify(data, null, 2),
  };
}
