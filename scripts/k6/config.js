// k6 測試配置文件

// 測試環境設定
export const TEST_CONFIG = {
  BASE_URL: __ENV.BASE_URL || "http://localhost:80",
  API_PREFIX: "",
  WS_URL: __ENV.WS_URL || "ws://localhost:80/ws",
  RESULTS_DIR: "test_results",
  // 增加預設 Header 供所有測試請求使用，解決 ENV=production 時的 VerifyOrigin 驗證問題
  DEFAULT_HEADERS: {
    Origin: "http://localhost:3000",
    "X-Loadtest-Bypass": __ENV.ALLOW_RATE_LIMIT_BYPASS || "0",
  },

  // 效能閾值
  THRESHOLDS: {
    http_req_duration: ["p(95)<200", "p(99)<500"],
    http_req_failed: ["rate<0.01"], // 1% 失敗率
    checks: ["rate>0.95"], // 95% check 成功率
  },

  // 測試階段配置
  SCENARIOS: {
    // 冒煙測試:最小負載,快速驗證功能
    smoke: [
      { duration: "20s", target: 1 }, // 20 秒內維持 1 個 VU
      { duration: "5s", target: 0 }, // 5 秒內降到 0 個 VU
    ],

    // WebSocket 廣播正確性測試：多 VU 同聚一室，驗證 Redis Pub/Sub 廣播
    // 10% Sender + 90% Listener，觀察廣播到達率與延遲
    ws_broadcast: [
      { duration: "1m",  target: 20  }, // 熱身：20 VU
      { duration: "2m",  target: 50  }, // 中等：50 VU (5 senders, 45 listeners)
      { duration: "3m",  target: 100 }, // 壓力：100 VU (10 senders, 90 listeners)
      { duration: "1m",  target: 0   }, // 收尾
    ],

    // 單體容量測試: 分級壓測 (100 -> 1000 VU)
    // 目的：建立本地基線、觀察趨勢與迴歸驗證
    // 注意：本地環境（Docker/OS）可能先於服務本身成為瓶頸，數字僅供趨勢參考
    monolith_capacity: [
      { duration: "2m", target: 100 }, // 階梯 1: 100 VU (功能驗證)
      { duration: "2m", target: 200 }, // 階梯 2: 200 VU (功能驗證)
      { duration: "3m", target: 500 }, // 階梯 3: 500 VU (趨勢觀察)
      { duration: "3m", target: 800 }, // 階梯 4: 800 VU (趨勢觀察)
      { duration: "5m", target: 1000 }, // 階梯 5: 1000 VU (本地壓力上限)
      { duration: "3m", target: 1000 }, // 維持 3 分鐘
      { duration: "2m", target: 0 }, // 結束
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
