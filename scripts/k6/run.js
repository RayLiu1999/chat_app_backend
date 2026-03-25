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
import http from "k6/http";
import { Counter, Rate } from 'k6/metrics';
import * as config from './config.js';
import smokeTest from './scenarios/smoke.js';
import monolithCapacityTest from './scenarios/monolith_capacity.js';
import wsBroadcastTest from './scenarios/ws_broadcast.js';
import { getAuthenticatedSessionWithOptions } from './scripts/common/auth.js';
import { logInfo, logSuccess, logError } from './scripts/common/logger.js';

// 自定義metrics用於即時監控
export const apiRequestCount = new Counter('api_requests_total');
export const apiSuccessRate = new Rate('api_success_rate');
export const wsConnectionCount = new Counter('ws_connections_total');

// 設定詳細日誌模式
const VERBOSE_MODE = __ENV.VERBOSE === '1';

// 選擇測試場景
const scenarioName = __ENV.SCENARIO || 'smoke';
const PREPARE_USERS = __ENV.PREPARE_USERS !== '0';
const PREPARE_USER_COUNT = parseInt(__ENV.PREPARE_USER_COUNT || '0', 10);
const scenarios = {
  smoke: smokeTest,
  monolith_capacity: monolithCapacityTest,
  ws_broadcast: wsBroadcastTest,
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
  
  const baseUrl = `${config.TEST_CONFIG.BASE_URL}${config.TEST_CONFIG.API_PREFIX}`;
  
  // 嘗試讀取預定義用戶數，若失敗則預設為 5
  let defaultPrepareCount = 5;
  try {
    const usersData = JSON.parse(open('./data/users.json'));
    defaultPrepareCount = (usersData && usersData.length) || 5;
  } catch (e) {
    // 忽略讀取錯誤，使用預設值
  }
  
  const prepareCount = Number.isInteger(PREPARE_USER_COUNT) && PREPARE_USER_COUNT > 0 ? PREPARE_USER_COUNT : defaultPrepareCount;

  // 1) 資料準備：預先建立/登入指定數量用戶，避免測試期產生註冊風暴
  const sessions = [];
  if (PREPARE_USERS) {
    console.log(`🧰 預先準備測試用戶: ${prepareCount} 位`);
    for (let index = 1; index <= prepareCount; index++) {
      try {
        const preparedSession = getAuthenticatedSessionWithOptions(baseUrl, {
          userIndex: index,
          registerIfMissing: true,
        });
        if (preparedSession && preparedSession.token) {
          sessions.push(preparedSession);
        }
      } catch (e) {
        console.error(`❌ 準備用戶 ${index} 失敗: ${e.message}`);
      }
    }
    console.log(`✅ 預備完成，可用 session: ${sessions.length}/${prepareCount}`);
  }

  // 2) 相容 fallback：若未啟用預備流程或全失敗，至少準備一組 session
  let session = null;
  if (sessions.length > 0) {
    session = sessions[0];
  } else {
    try {
      session = getAuthenticatedSessionWithOptions(baseUrl, {
        userIndex: 1,
        registerIfMissing: true,
      });
      if (session && session.token) {
        console.log(`🔑 身份驗證成功，Token: ${session.token.substring(0, 10)}...`);
      } else {
        console.warn(`⚠️ 身份驗證未返回有效 Session`);
      }
    } catch (e) {
      console.error(`❌ 身份驗證失敗: ${e.message}`);
    }
  }

  // 3) ws_broadcast 場景：額外建立共用廣播頻道
  let broadcastChannelId = null;
  if (scenarioName === 'ws_broadcast') {
    // ws_broadcast 需要大量 VU，確保至少準備 100 個 session
    const broadcastUserCount = Math.max(sessions.length, 100);
    if (sessions.length < broadcastUserCount) {
      for (let i = sessions.length + 1; i <= broadcastUserCount; i++) {
        try {
          const s = getAuthenticatedSessionWithOptions(baseUrl, {
            userIndex: i,
            registerIfMissing: true,
          });
          if (s && s.token) sessions.push(s);
        } catch (_) {}
      }
    }
    console.log(`🔌 [Broadcast Setup] 準備 ${sessions.length} 個 session`);
    // 還原 session[0] 的 CSRF cookie（準備 100 個 session 後 jar 被最後一個覆蓋）
    if (session && session.csrfToken) {
      const jar = http.cookieJar();
      jar.set(config.TEST_CONFIG.BASE_URL, "csrf_token", session.csrfToken, { path: "/" });
    }
    const broadcastSetup = setupBroadcastChannel(baseUrl, session);
    if (broadcastSetup) {
      broadcastChannelId = broadcastSetup.channelId;
      console.log(`✅ [Broadcast Setup] 共用 Channel 建立成功: ${broadcastChannelId}`);

      // 讓所有 session 加入 server，否則非成員無法 join_room
      const joinUrl = `${baseUrl}/servers/${broadcastSetup.serverId}/join`;
      let joinedCount = 0;
      for (let i = 1; i < sessions.length; i++) {
        const s = sessions[i];
        if (!s || !s.token) continue;
        // 還原每個 session 的 CSRF cookie
        if (s.csrfToken) {
          const jar = http.cookieJar();
          jar.set(config.TEST_CONFIG.BASE_URL, "csrf_token", s.csrfToken, { path: "/" });
        }
        const res = http.post(joinUrl, null, {
          headers: {
            ...s.headers,
            'X-CSRF-TOKEN': s.csrfToken || '',
            'Origin': 'http://localhost:3000',
          },
        });
        if (res.status === 200 || res.status === 409) {
          joinedCount++;
        }
      }
      console.log(`✅ [Broadcast Setup] ${joinedCount}/${sessions.length - 1} 個用戶加入 Server`);
    } else {
      console.error(`❌ [Broadcast Setup] 共用 Channel 建立失敗，測試可能無法正常執行`);
    }
  }

  return {
    startTime: Date.now(),
    scenario: scenarioName,
    verbose: VERBOSE_MODE,
    sessions: sessions,
    session: session,
    config: config.TEST_CONFIG,
    broadcastChannelId: broadcastChannelId,
  };
}

/**
 * 為 ws_broadcast 場景在 setup 完成後建立共用廣播頻道
 * 回傳 { serverId, channelId } 或 null
 */
function setupBroadcastChannel(baseUrl, session) {
  if (!session || !session.token) return null;

  const commonHeaders = {
    ...session.headers,
    'Content-Type': 'application/json',
    'Origin': 'http://localhost:3000',
  };

  // 1) 建立專屬測試 Server（multipart/form-data 手工拼接）
  let serverId = null;
  let channelId = null;

  try {
    const formHeaders = {
      ...session.headers,
      'Origin': 'http://localhost:3000',
    };
    // multipart/form-data 需要手工拼接
    const boundary = `----k6Boundary${Date.now()}`;
    const serverName = `BroadcastTestServer-${Date.now()}`;
    const body = [
      `--${boundary}`,
      'Content-Disposition: form-data; name="name"',
      '',
      serverName,
      `--${boundary}`,
      'Content-Disposition: form-data; name="description"',
      '',
      'k6 WS broadcast test server',
      `--${boundary}`,
      'Content-Disposition: form-data; name="is_public"',
      '',
      'false',
      `--${boundary}--`,
    ].join('\r\n');

    const csrfHeaders = {};
    if (session.csrfToken) {
      csrfHeaders['X-CSRF-TOKEN'] = session.csrfToken;
    }

    const res = http.post(`${baseUrl}/servers`, body, {
      headers: {
        ...formHeaders,
        ...csrfHeaders,
        'Content-Type': `multipart/form-data; boundary=${boundary}`,
      },
    });

    if (res.status === 200) {
      const data = res.json('data');
      serverId = data && data.id;
      console.log(`🏠 [Broadcast Setup] 建立 Server: ${serverId}`);
    } else {
      console.error(`❌ [Broadcast Setup] 建立 Server 失敗: ${res.status} ${res.body}`);
      return null;
    }
  } catch (e) {
    console.error(`❌ [Broadcast Setup] Server 建立例外: ${e.message}`);
    return null;
  }

  if (!serverId) return null;

  // 2) 在 Server 建立頻道
  try {
    const csrfHeaders = {};
    if (session.csrfToken) {
      csrfHeaders['X-CSRF-TOKEN'] = session.csrfToken;
    }
    const res = http.post(
      `${baseUrl}/servers/${serverId}/channels`,
      JSON.stringify({ name: 'broadcast-test', type: 'text' }),
      {
        headers: {
          ...commonHeaders,
          ...csrfHeaders,
        },
      }
    );

    if (res.status === 200 || res.status === 201) {
      channelId = res.json('data.id');
      console.log(`📺 [Broadcast Setup] 建立 Channel: ${channelId}`);
    } else {
      // Fallback：嘗試抓現有 channel
      const chRes = http.get(`${baseUrl}/servers/${serverId}/channels`, {
        headers: { ...commonHeaders },
      });
      if (chRes.status === 200) {
        const chData = chRes.json('data');
        if (Array.isArray(chData) && chData.length > 0) {
          channelId = chData[0].id;
          console.log(`📺 [Broadcast Setup] 使用現有 Channel: ${channelId}`);
        }
      }
    }
  } catch (e) {
    console.error(`❌ [Broadcast Setup] Channel 建立例外: ${e.message}`);
  }

  return channelId ? { serverId, channelId } : null;
}

// 主執行函數
export default function (data) {
  const iterationStart = Date.now();

  let session = data.session || null;
  if (data.sessions && data.sessions.length > 0) {
    const idx = (__VU - 1) % data.sessions.length;
    session = data.sessions[idx];
  }
  
  // 同步 CSRF Cookie 到當前 VU 的 Cookie Jar (解決 setup() 資料不會自動同步 Cookie 的問題)
  if (session && session.csrfToken) {
    const jar = http.cookieJar();
    jar.set(data.config.BASE_URL, "csrf_token", session.csrfToken, { path: "/" });
  }

  logInfo(`開始執行迭代 - 場景: ${data?.scenario || scenarioName}`);
  
  try {
    if (data.scenario === 'ws_broadcast') {
      // ws_broadcast 需要 broadcastChannelId，傳入完整的 data 物件
      scenarios[scenarioName](data.config, data);
    } else {
      // 其他場景維持原有介面：(config, session)
      scenarios[scenarioName](data.config, session);
    }
    
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

  // ws_broadcast 場景專屬指標
  if (scenario === 'ws_broadcast') {
    const sent = getMetricValue('ws_broadcast_sent', 'count');
    const received = getMetricValue('ws_broadcast_received', 'count');
    const joinSuccess = getMetricValue('ws_room_join_success', 'count');
    const joinFailed = getMetricValue('ws_room_join_failed', 'count');
    const latencyTotal = getMetricValue('ws_broadcast_latency_ms_total', 'count');
    const avgLatency = sent > 0 ? (latencyTotal / received).toFixed(2) : 'N/A';

    report += '\n## WebSocket 廣播測試\n\n';
    report += `* **訊息發送總數 (Sender):** ${sent}\n`;
    report += `* **訊息接收總數 (所有 Listener):** ${received}\n`;
    report += `* **房間加入成功:** ${joinSuccess}\n`;
    report += `* **房間加入失敗:** ${joinFailed}\n`;
    report += `* **平均廣播延遲:** ${avgLatency}ms\n`;
    if (sent > 0 && received > 0) {
      report += `* **廣播倍率 (接收/發送):** ${(received / sent).toFixed(1)}x (= 平均每則廣播到達的 VU 數)\n`;
    }
  }

  return {
    stdout: report,
    [`${outputDir}/${now}_${scenario}_summary.md`]: report,
    [`${outputDir}/${now}_${scenario}_results.json`]: JSON.stringify(data, null, 2),
  };
}
