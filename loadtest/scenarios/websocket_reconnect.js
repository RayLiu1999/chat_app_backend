/**
 * WebSocket 斷線重連測試場景 (Reconnection Test)
 *
 * 目的：測試 WebSocket 斷線重連的穩定性和速度
 *
 * 測試場景：
 * 1. 單次重連：斷線後重新連線，驗證狀態恢復
 * 2. 重連風暴：大量用戶同時斷線重連
 * 3. 頻繁重連：模擬極差網路環境
 *
 * 使用方法：
 * k6 run run.js --env SCENARIO=websocket_reconnect
 * k6 run run.js --env SCENARIO=websocket_reconnect --env RECONNECT_TYPE=storm
 * k6 run run.js --env SCENARIO=websocket_reconnect --env RECONNECT_TYPE=frequent
 */

import { check, sleep, group } from "k6";
import ws from "k6/ws";
import { Counter, Trend, Rate, Gauge } from "k6/metrics";
import { getAuthenticatedSession } from "../scripts/common/auth.js";
import {
  logInfo,
  logError,
  logSuccess,
  logGroupStart,
  logGroupEnd,
} from "../scripts/common/logger.js";

// 自定義重連指標
const wsReconnectTime = new Trend("ws_reconnect_time");
const wsReconnectSuccess = new Rate("ws_reconnect_success");
const wsReconnectAttempts = new Counter("ws_reconnect_attempts");
const wsStateRecoverySuccess = new Rate("ws_state_recovery_success");
const wsReconnectErrors = new Counter("ws_reconnect_errors");
const wsActiveReconnections = new Gauge("ws_active_reconnections");

// 測試類型
const RECONNECT_TYPE = __ENV.RECONNECT_TYPE || "standard";

export default function (config) {
  const vuNumber = __VU;
  const iteration = __ITER;

  logGroupStart(`Reconnection Test - VU ${vuNumber}`);
  logInfo(`🔄 重連測試類型: ${RECONNECT_TYPE}`);

  // 取得認證會話
  const session = getAuthenticatedSession(
    `${config.BASE_URL}${config.API_PREFIX}`
  );

  if (!session) {
    logError("認證失敗，無法進行重連測試");
    return;
  }

  // 根據測試類型選擇不同的測試策略
  switch (RECONNECT_TYPE) {
    case "storm":
      testReconnectStorm(config, session);
      break;
    case "frequent":
      testFrequentReconnect(config, session);
      break;
    default:
      testStandardReconnect(config, session);
  }

  logGroupEnd(`Reconnection Test - VU ${vuNumber}`, Date.now());
}

/**
 * 標準重連測試
 * 模擬正常使用情境下的斷線重連
 */
function testStandardReconnect(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const testRoomId = `reconnect_room_${__VU % 5}`; // 5 個測試房間

  logInfo(`📍 標準重連測試 - 房間: ${testRoomId}`);

  // 第一次連線
  let firstConnectionState = null;

  group("Phase 1: 初始連線", () => {
    firstConnectionState = establishConnection(wsUrl, testRoomId, session);
  });

  if (!firstConnectionState || !firstConnectionState.success) {
    logError("初始連線失敗，跳過重連測試");
    return;
  }

  // 模擬斷線（等待一段時間模擬網路中斷）
  group("Phase 2: 模擬斷線", () => {
    const disconnectDuration = 3 + Math.random() * 2; // 3-5 秒斷線
    logInfo(`⏸️  模擬斷線 ${disconnectDuration.toFixed(1)} 秒...`);
    sleep(disconnectDuration);
  });

  // 重新連線
  group("Phase 3: 重新連線", () => {
    wsReconnectAttempts.add(1);
    wsActiveReconnections.add(1);

    const reconnectStart = Date.now();
    const reconnectState = establishConnection(wsUrl, testRoomId, session, true);
    const reconnectTime = Date.now() - reconnectStart;

    wsReconnectTime.add(reconnectTime);
    wsActiveReconnections.add(-1);

    if (reconnectState && reconnectState.success) {
      wsReconnectSuccess.add(1);
      logSuccess(`重連成功`, 101, reconnectTime);

      // 驗證狀態恢復
      const stateRecovered = verifyStateRecovery(
        firstConnectionState,
        reconnectState
      );
      wsStateRecoverySuccess.add(stateRecovered ? 1 : 0);

      if (stateRecovered) {
        logInfo("✅ 狀態恢復成功");
      } else {
        logError("❌ 狀態恢復失敗");
      }
    } else {
      wsReconnectSuccess.add(0);
      wsReconnectErrors.add(1);
      logError(`重連失敗 (耗時: ${reconnectTime}ms)`);
    }
  });

  // 重連後發送測試訊息
  group("Phase 4: 重連後功能驗證", () => {
    const verifyState = establishConnection(wsUrl, testRoomId, session);
    if (verifyState && verifyState.success) {
      logInfo("✅ 重連後功能正常");
    }
  });
}

/**
 * 重連風暴測試
 * 模擬大量用戶同時斷線重連（如服務器重啟）
 */
function testReconnectStorm(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const stormRoomId = "storm_test_room"; // 所有用戶同一房間

  logInfo(`🌪️  重連風暴測試 - VU ${__VU}`);

  // 建立初始連線
  group("Storm: 初始連線", () => {
    const state = establishConnection(wsUrl, stormRoomId, session);
    if (!state || !state.success) {
      logError("初始連線失敗");
      return;
    }
  });

  // 同步等待（模擬所有用戶同時斷線）
  group("Storm: 同步斷線", () => {
    // 使用固定時間點讓所有 VU 同時開始重連
    const syncDelay = 5 - (__VU % 5); // 0-4 秒的隨機延遲，模擬略微錯開
    sleep(syncDelay);
    logInfo(`⚡ VU ${__VU} 開始重連風暴`);
  });

  // 同時重連
  group("Storm: 大量重連", () => {
    wsReconnectAttempts.add(1);

    const reconnectStart = Date.now();
    const state = establishConnection(wsUrl, stormRoomId, session, true);
    const reconnectTime = Date.now() - reconnectStart;

    wsReconnectTime.add(reconnectTime);

    const success = state && state.success;
    wsReconnectSuccess.add(success ? 1 : 0);

    if (success) {
      logSuccess(`風暴重連成功`, 101, reconnectTime);
    } else {
      wsReconnectErrors.add(1);
      logError(`風暴重連失敗 (${reconnectTime}ms)`);
    }
  });

  // 穩定性驗證
  group("Storm: 穩定性驗證", () => {
    sleep(5); // 等待系統穩定
    const state = establishConnection(wsUrl, stormRoomId, session);
    if (state && state.success) {
      logInfo("✅ 風暴後系統穩定");
    } else {
      logError("❌ 風暴後系統不穩定");
    }
  });
}

/**
 * 頻繁重連測試
 * 模擬極差網路環境下的頻繁斷線重連
 */
function testFrequentReconnect(config, session) {
  const wsUrl = `${config.WS_URL}?token=${session.token}`;
  const frequentRoomId = `frequent_room_${__VU % 3}`;

  logInfo(`🔁 頻繁重連測試 - VU ${__VU}`);

  const reconnectCycles = 5; // 執行 5 次重連循環
  let successCount = 0;
  let totalReconnectTime = 0;

  for (let cycle = 1; cycle <= reconnectCycles; cycle++) {
    group(`Frequent: 循環 ${cycle}/${reconnectCycles}`, () => {
      // 建立連線
      const connectStart = Date.now();
      const state = establishConnection(wsUrl, frequentRoomId, session, cycle > 1);
      const connectTime = Date.now() - connectStart;

      if (state && state.success) {
        successCount++;
        totalReconnectTime += connectTime;
        logInfo(`✅ 循環 ${cycle} 連線成功 (${connectTime}ms)`);

        // 短暫保持連線
        const holdTime = 5 + Math.random() * 5; // 5-10 秒
        sleep(holdTime);
      } else {
        logError(`❌ 循環 ${cycle} 連線失敗`);
      }

      wsReconnectAttempts.add(1);
      wsReconnectSuccess.add(state && state.success ? 1 : 0);
      if (state && state.success) {
        wsReconnectTime.add(connectTime);
      } else {
        wsReconnectErrors.add(1);
      }

      // 斷線間隔
      if (cycle < reconnectCycles) {
        const disconnectTime = 2 + Math.random() * 3; // 2-5 秒
        logInfo(`⏸️  斷線 ${disconnectTime.toFixed(1)} 秒...`);
        sleep(disconnectTime);
      }
    });
  }

  // 統計結果
  const avgReconnectTime =
    successCount > 0 ? totalReconnectTime / successCount : 0;
  logInfo(`📊 頻繁重連統計:`);
  logInfo(`   成功率: ${((successCount / reconnectCycles) * 100).toFixed(1)}%`);
  logInfo(`   平均重連時間: ${avgReconnectTime.toFixed(0)}ms`);
}

/**
 * 建立 WebSocket 連線
 */
function establishConnection(wsUrl, roomId, session, isReconnect = false) {
  const connectionState = {
    success: false,
    roomJoined: false,
    messagesReceived: 0,
    connectionTime: 0,
  };

  const connectionStart = Date.now();
  const logPrefix = isReconnect ? "🔄 重連" : "🔌 連線";

  try {
    const response = ws.connect(
      wsUrl,
      {
        headers: {
          Authorization: `Bearer ${session.token}`,
        },
        tags: { test_type: "reconnect" },
      },
      function (socket) {
        connectionState.connectionTime = Date.now() - connectionStart;
        connectionState.success = true;

        logInfo(`${logPrefix}成功 (${connectionState.connectionTime}ms)`);

        socket.on("open", () => {
          // 加入房間
          socket.send(
            JSON.stringify({
              action: "join_room",
              room_id: roomId,
            })
          );
        });

        socket.on("message", (data) => {
          connectionState.messagesReceived++;
          try {
            const message = JSON.parse(data);
            if (
              message.action === "status" &&
              message.message &&
              message.message.includes("加入房間成功")
            ) {
              connectionState.roomJoined = true;
              logInfo(`📥 加入房間成功: ${roomId}`);
            }
          } catch (e) {
            // 忽略解析錯誤
          }
        });

        socket.on("error", (e) => {
          logError(`WebSocket 錯誤: ${e.error ? e.error() : e}`);
          connectionState.success = false;
        });

        // 短暫保持連線以確認狀態
        sleep(2);

        // 發送測試訊息
        socket.send(
          JSON.stringify({
            action: "send_message",
            room_id: roomId,
            content: `${isReconnect ? "重連" : "連線"}測試訊息 from VU ${__VU}`,
            message_type: "text",
          })
        );

        sleep(1);
        socket.close();
      }
    );

    check(response, {
      [`${logPrefix} WebSocket 連線成功`]: (r) => r && r.status === 101,
    });
  } catch (e) {
    logError(`${logPrefix}失敗: ${e.message}`);
    connectionState.success = false;
  }

  return connectionState;
}

/**
 * 驗證狀態恢復
 */
function verifyStateRecovery(previousState, currentState) {
  // 基本驗證：重連後能正常加入房間
  if (!currentState.success) {
    return false;
  }

  // 驗證房間狀態恢復
  if (previousState.roomJoined && !currentState.roomJoined) {
    logError("房間狀態未恢復");
    return false;
  }

  return true;
}
