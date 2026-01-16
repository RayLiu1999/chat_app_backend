// k6 測試配置文件

// 測試環境設定
export const TEST_CONFIG = {
  BASE_URL: __ENV.BASE_URL || "http://localhost:80",
  API_PREFIX: "",
  WS_URL: __ENV.WS_URL || "ws://localhost:80/ws",
  RESULTS_DIR: "test_results",

  // 測試用戶配置（根據 API.md 規範）
  TEST_USERS: [
    {
      username: "testuser1",
      nickname: "Test User 1",
      email: "testuser1@example.com",
      password: "Password123!",
    },
    {
      username: "testuser2",
      nickname: "Test User 2",
      email: "testuser2@example.com",
      password: "Password123!",
    },
    {
      username: "testuser3",
      nickname: "Test User 3",
      email: "testuser3@example.com",
      password: "Password123!",
    },
    {
      username: "testuser4",
      nickname: "Test User 4",
      email: "testuser4@example.com",
      password: "Password123!",
    },
    {
      username: "testuser5",
      nickname: "Test User 5",
      email: "testuser5@example.com",
      password: "Password123!",
    },
  ],

  // 測試房間配置
  TEST_ROOMS: [
    { id: "test_room_001", name: "測試房間 1", type: "channel" },
    { id: "test_room_002", name: "測試房間 2", type: "channel" },
    { id: "test_room_003", name: "測試房間 3", type: "dm" },
  ],

  // 效能閾值
  THRESHOLDS: {
    http_req_duration: ["p(95)<200", "p(99)<500"],
    http_req_failed: ["rate<0.01"], // 1% 失敗率
    checks: ["rate>0.95"], // 95% check 成功率
    ws_connect_time: ["p(95)<300"],
    ws_message_send_duration: ["p(95)<50"],
  },

  // 測試階段配置
  SCENARIOS: {
    // 冒煙測試：最小負載，快速驗證功能
    smoke: [
      { duration: "20s", target: 1 }, // 20 秒內維持 1 個 VU
      { duration: "5s", target: 0 }, // 5 秒內降到 0 個 VU
    ],

    // 輕量測試：小規模負載測試
    light: [
      { duration: "30s", target: 5 }, // 30 秒內提升到 5 個 VU
      { duration: "1m", target: 10 }, // 1 分鐘內提升到 10 個 VU
      { duration: "30s", target: 0 }, // 30 秒內降到 0 個 VU
    ],

    // 中量測試：中等規模負載測試
    medium: [
      { duration: "1m", target: 20 }, // 1 分鐘內提升到 20 個 VU
      { duration: "3m", target: 50 }, // 3 分鐘內提升到 50 個 VU
      { duration: "1m", target: 0 }, // 1 分鐘內降到 0 個 VU
    ],

    // 重量測試：大規模負載測試
    heavy: [
      { duration: "2m", target: 50 }, // 2 分鐘內提升到 50 個 VU
      { duration: "5m", target: 100 }, // 5 分鐘內提升到 100 個 VU
      { duration: "2m", target: 200 }, // 2 分鐘內提升到 200 個 VU
      { duration: "1m", target: 0 }, // 1 分鐘內降到 0 個 VU
    ],

    // ========== WebSocket 專用壓力測試 ==========

    // WebSocket 連線壓力測試：測試大量並發連線
    websocket_stress: [
      { duration: "30s", target: 50 }, // 30 秒內建立 50 個連線
      { duration: "2m", target: 100 }, // 2 分鐘內提升到 100 個連線
      { duration: "3m", target: 150 }, // 3 分鐘內提升到 150 個連線
      { duration: "5m", target: 150 }, // 維持 150 個連線 5 分鐘
      { duration: "1m", target: 0 }, // 1 分鐘內關閉所有連線
    ],

    // WebSocket 峰值測試：突然湧入大量連線
    websocket_spike: [
      { duration: "10s", target: 10 }, // 預熱：10 秒到 10 個 VU
      { duration: "10s", target: 200 }, // 峰值：10 秒內暴增到 200 個 VU
      { duration: "1m", target: 200 }, // 維持峰值 1 分鐘
      { duration: "30s", target: 10 }, // 回落到 10 個 VU
      { duration: "10s", target: 0 }, // 關閉
    ],

    // WebSocket 浸泡測試：長時間穩定性測試
    websocket_soak: [
      { duration: "2m", target: 50 }, // 2 分鐘升到 50 個 VU
      { duration: "1h", target: 50 }, // 維持 50 個 VU 運行 1 小時
      { duration: "2m", target: 0 }, // 2 分鐘內關閉
    ],

    // WebSocket 階梯測試：逐步增加負載找出系統極限
    websocket_stress_ladder: [
      { duration: "2m", target: 50 }, // 第一階：50 個連線
      { duration: "2m", target: 50 }, // 維持
      { duration: "2m", target: 100 }, // 第二階：100 個連線
      { duration: "2m", target: 100 }, // 維持
      { duration: "2m", target: 200 }, // 第三階：200 個連線
      { duration: "2m", target: 200 }, // 維持
      { duration: "2m", target: 300 }, // 第四階：300 個連線
      { duration: "2m", target: 300 }, // 維持
      { duration: "2m", target: 500 }, // 第五階：500 個連線
      { duration: "2m", target: 500 }, // 維持
      { duration: "2m", target: 0 }, // 關閉
    ],

    // WebSocket 斷線重連測試：測試網路不穩定情況
    websocket_reconnect: [
      { duration: "1m", target: 20 },  // 1 分鐘升到 20 個 VU
      { duration: "3m", target: 50 },  // 3 分鐘升到 50 個 VU
      { duration: "2m", target: 50 },  // 維持 50 個 VU
      { duration: "1m", target: 0 },   // 1 分鐘內關閉
    ],
  },
};

// API 端點（根據實際的 routes.go）
export const API_ENDPOINTS = {
  // 認證相關
  REGISTER: "/register",
  LOGIN: "/login",
  LOGOUT: "/logout",
  REFRESH: "/refresh_token",

  // 健康檢查
  HEALTH: "/health",

  // 用戶相關
  USER: "/user",
  USER_PROFILE: "/user/profile",

  // 伺服器相關
  SERVERS: "/servers",
  SERVER_SEARCH: "/servers/search",
  SERVER_BY_ID: (id) => `/servers/${id}`,
  SERVER_DETAIL: (id) => `/servers/${id}/detail`,
  SERVER_CHANNELS: (id) => `/servers/${id}/channels`,
  SERVER_JOIN: (id) => `/servers/${id}/join`,
  SERVER_LEAVE: (id) => `/servers/${id}/leave`,

  // 頻道相關
  CHANNEL_CREATE: (serverId) => `/servers/${serverId}/channels`,
  CHANNEL_BY_ID: (channelId) => `/channels/${channelId}`,
  CHANNEL_MESSAGES: (channelId) => `/channels/${channelId}/messages`,

  // 好友相關
  FRIENDS: "/friends",
  FRIENDS_PENDING: "/friends/pending",
  FRIENDS_BLOCKED: "/friends/blocked",

  // 私聊相關
  DM_ROOMS: "/dm_rooms",
  DM_ROOM_MESSAGES: (roomId) => `/dm_rooms/${roomId}/messages`,

  // 檔案上傳
  FILE_UPLOAD: "/upload/file",
};

// WebSocket 訊息類型（根據後端 websocket_handler.go）
export const WS_MESSAGE_TYPES = {
  // 客戶端發送的動作（使用 action 而不是 type）
  JOIN_ROOM: "join_room",
  LEAVE_ROOM: "leave_room",
  SEND_MESSAGE: "send_message",
  PING: "ping",

  // 伺服器回應的動作
  ERROR: "error",
  STATUS: "status",
};
