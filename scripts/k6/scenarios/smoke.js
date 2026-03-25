/**
 * 冒煙測試場景 (Smoke Test)
 *
 * 目的：驗證所有核心 API 和 WebSocket 功能是否正常工作
 * 特點：
 * - 最小負載（1-2 個 VU）
 * - 快速執行（30 秒內完成）
 * - 覆蓋所有關鍵 API 端點
 * - 使用真實創建的資源進行測試
 */
import { group, sleep, check } from "k6";
import http from "k6/http";
import { randomSleep } from "../scripts/common/utils.js";
import { getAuthenticatedSession } from "../scripts/common/auth.js";
import apiAuth from "../scripts/api/auth.js";
import apiServers from "../scripts/api/servers.js";
import apiFriends from "../scripts/api/friends.js";
import apiChat from "../scripts/api/chat.js";
import wsConnect from "../scripts/websocket/connection.js";
import { joinRoom, leaveRoom } from "../scripts/websocket/rooms.js";
import { sendMessage } from "../scripts/websocket/messaging.js";
import {
  logInfo,
  logError,
  logGroupStart,
  logGroupEnd,
} from "../scripts/common/logger.js";

export default function (config, session) {
  const testStartTime = Date.now();
  const baseUrl = `${config.BASE_URL}${config.API_PREFIX}`;

  logGroupStart("Smoke Test - 冒煙測試");
  logInfo(`測試環境: ${baseUrl}`);

  // ==================== 階段 1: 認證測試 ====================
  group("Phase 1: Authentication", function () {
    logInfo("開始認證階段測試");
    apiAuth(baseUrl);
    randomSleep(0.5, 1);
  });

  // ==================== 階段 2: 驗證已認證會話 ====================
  group("Phase 2: Verify Authenticated Session", function () {
    if (!session) {
      logError("⚠️  無法建立認證會話，跳過需要認證的測試");
      return;
    }

    // 還原當前 session 的 CSRF cookie，避免 Phase 1 的登入流程覆蓋 cookie jar。
    if (session.csrfToken) {
      const jar = http.cookieJar();
      jar.set(config.BASE_URL, "csrf_token", session.csrfToken, { path: "/" });
    }

    logInfo(`✅ 會話建立成功，用戶: ${session.user.email}`);
    logInfo(`✅ Access Token: ${session.token.substring(0, 20)}...`);
    if (session.csrfToken) {
      logInfo(`✅ CSRF Token: ${session.csrfToken.substring(0, 20)}...`);
    }
  });

  if (!session) {
    logGroupEnd("Smoke Test - 冒煙測試 (部分失敗)", testStartTime);
    return;
  }

  // ==================== 階段 3: 伺服器管理測試 ====================
  let serverData;

  group("Phase 3: Server Management", function () {
    logInfo("開始伺服器管理測試");
    serverData = apiServers(baseUrl, session);

    if (serverData && serverData.serverId) {
      logInfo(`✅ 伺服器測試完成，伺服器 ID: ${serverData.serverId}`);
    }
    if (serverData && serverData.channelId) {
      logInfo(`✅ 頻道創建完成，頻道 ID: ${serverData.channelId}`);
    }

    randomSleep(0.5, 1);
  });

  // ==================== 階段 4: 好友系統測試 ====================
  group("Phase 4: Friend Management", function () {
    logInfo("開始好友系統測試");
    apiFriends(baseUrl, session);
    randomSleep(0.5, 1);
  });

  // ==================== 階段 5: 聊天功能測試 ====================
  let chatData;

  group("Phase 5: Chat Management", function () {
    logInfo("開始聊天功能測試");
    chatData = apiChat(baseUrl, session);

    if (chatData && chatData.dmRoomId) {
      logInfo(`✅ 找到 DM 房間，房間 ID: ${chatData.dmRoomId}`);
    }

    randomSleep(0.5, 1);
  });

  // ==================== 階段 6: WebSocket 連線測試 ====================
  group("Phase 6: WebSocket Connection Test", function () {
    logInfo("開始 WebSocket 連線測試");

    if (!session.token) {
      logError("❌ 缺少 Access Token，跳過 WebSocket 測試");
      return;
    }

    try {
      // 測試 1: 使用真實創建的頻道進行測試
      if (serverData && serverData.channelId) {
        group("Test Channel WebSocket", function () {
          logInfo(`測試頻道 WebSocket 連線: ${serverData.channelId}`);

          // ⭐ 重構：handler 只發送訊息，不檢查結果
          const result = wsConnect(
            config.WS_URL,
            session.token,
            function (socket) {
              let messagesSent = 0;

              // 步驟 1: 加入頻道
              logInfo(`步驟 1: 加入頻道 ${serverData.channelId}`);
              joinRoom(socket, serverData.channelId, "channel");
              messagesSent++;
              sleep(3);

              // 步驟 2: 發送訊息
              logInfo("步驟 2: 發送測試訊息");
              sendMessage(socket, serverData.channelId, "channel");
              messagesSent++;
              sleep(2);

              // 步驟 3: 測試 Ping/Pong
              logInfo("步驟 3: 測試 Ping/Pong");
              socket.send(JSON.stringify({ action: "ping", data: {} }));
              messagesSent++;
              sleep(1);

              // 步驟 4: 離開頻道
              logInfo("步驟 4: 離開頻道");
              leaveRoom(socket, serverData.channelId, "channel");
              messagesSent++;
              sleep(2);

              return { messagesSent };
            },
            30
          );

          // ⭐ 在 wsConnect 返回後檢查結果
          logInfo(`📊 收到 ${result.receivedMessages.length} 條訊息`);

          // 使用 check 驗證 WebSocket 訊息
          check(result.messageStates, {
            "WS Channel: room_joined received": (s) => s.room_joined === true,
            "WS Channel: message_sent received": (s) => s.message_sent === true,
            "WS Channel: pong received": (s) => s.pong === true,
            "WS Channel: room_left received": (s) => s.room_left === true,
          });

          // 記錄詳細結果
          if (result.messageStates.room_joined) {
            logInfo("✅ 成功加入頻道");
          } else {
            logError("❌ 未收到 room_joined 回應");
          }

          if (result.messageStates.message_sent) {
            logInfo("✅ 訊息發送成功");
          } else {
            logError("❌ 未收到 message_sent 確認");
          }

          if (result.messageStates.pong) {
            logInfo("✅ Ping/Pong 正常");
          } else {
            logError("❌ 未收到 pong 回應");
          }

          if (result.messageStates.room_left) {
            logInfo("✅ 成功離開頻道");
          } else {
            logError("❌ 未收到 room_left 確認");
          }

          if (result.success) {
            logInfo("✅ 頻道 WebSocket 測試完成");
          } else {
            logError("❌ 頻道 WebSocket 測試失敗");
          }
        });

        sleep(1); // 兩個測試之間的延遲
      } else {
        logInfo("⚠️  沒有可用的頻道 ID，跳過頻道 WebSocket 測試");
      }

      // 測試 2: 使用 DM 房間進行測試（如果存在）
      if (chatData && chatData.dmRoomId) {
        group("Test DM WebSocket", function () {
          logInfo(`測試 DM WebSocket 連線: ${chatData.dmRoomId}`);

          // ⭐ 重構：handler 只發送訊息
          const result = wsConnect(
            config.WS_URL,
            session.token,
            function (socket) {
              let messagesSent = 0;

              // 加入 DM 房間
              logInfo(`加入 DM 房間 ${chatData.dmRoomId}`);
              joinRoom(socket, chatData.dmRoomId, "dm");
              messagesSent++;
              sleep(3);

              return { messagesSent };
            },
            15
          );

          // ⭐ 在外部檢查結果
          check(result.messageStates, {
            "WS DM: room_joined or error received": (s) =>
              s.room_joined === true || s.error === true,
          });

          if (result.messageStates.room_joined || result.messageStates.error) {
            logInfo("✅ DM 房間測試完成");
          } else {
            logError("❌ DM 房間未收到回應");
          }

          if (result.success) {
            logInfo("✅ DM WebSocket 測試完成");
          }
        });
      } else {
        logInfo("⚠️  沒有可用的 DM 房間 ID，跳過 DM WebSocket 測試");
      }

      logInfo("✅ WebSocket 測試階段完成");
    } catch (e) {
      logError("WebSocket 測試執行異常", e.message);
    }
  });

  const totalDuration = Date.now() - testStartTime;
  logGroupEnd("Smoke Test - 冒煙測試完成", testStartTime);
  logInfo(
    `✅ 冒煙測試執行完畢，總耗時: ${(totalDuration / 1000).toFixed(2)} 秒`
  );
}
