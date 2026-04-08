/**
 * WebSocket 廣播正確性測試 (WS Broadcast Test)
 *
 * 測試目標：驗證 Redis Pub/Sub 廣播機制的正確性
 *
 * 場景設計：
 * - 所有 VU 同時加入同一個固定的 Channel
 * - VU 角色依 VU index 決定：
 *   - 前 SENDER_RATIO * VU 數 的 VU 為 Sender（發訊息）
 *   - 其餘為 Listener（純接收，驗證廣播是否到達）
 * - Sender 每 SEND_INTERVAL_MS 發送一則訊息
 * - Listener 計算收到的 new_message 數量
 *
 * 關鍵指標：
 * - ws_broadcast_sent:     Sender 發出的訊息總數
 * - ws_broadcast_received: 所有 VU 收到廣播的次數（N VU 收到 = N 次）
 * - ws_broadcast_missed:   Sender 預期每個 Listener 都收到，若超時未收到則計 miss
 * - ws_room_join_success:  加入房間成功率
 */
import ws from "k6/ws";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import http from "k6/http";
import { getAuthenticatedSessionWithOptions } from "../scripts/common/auth.js";
import { logInfo, logError } from "../scripts/common/logger.js";

// ── 自訂 Metrics ──────────────────────────────────────────
export const wsBroadcastSent = new Counter("ws_broadcast_sent");
export const wsBroadcastReceived = new Counter("ws_broadcast_received");
export const wsRoomJoinSuccess = new Counter("ws_room_join_success");
export const wsRoomJoinFailed = new Counter("ws_room_join_failed");
export const wsBroadcastLatency = new Counter("ws_broadcast_latency_ms_total");

// ── 場景常數 ─────────────────────────────────────────────
const SENDER_RATIO = 0.1; // 10% VU 為 sender
const SOAK_DURATION_MS = 3 * 60 * 1000; // 每個 VU 掛網 3 分鐘
const SEND_INTERVAL_MS = 3000; // Sender 每 3 秒發一則訊息
const ROOM_JOIN_TIMEOUT_MS = 5000;

/**
 * 廣播測試主函數
 * @param {Object} config  - TEST_CONFIG
 * @param {Object} setupData - setup() 回傳的資料 (含 channelId, sessions)
 */
export default function (config, setupData) {
  if (!setupData || !setupData.broadcastChannelId) {
    logError(`[Broadcast] setup 未提供 broadcastChannelId，跳過`);
    sleep(2);
    return;
  }

  const { broadcastChannelId, sessions } = setupData;

  // ── 取得本 VU 的 session ──────────────────────────────
  let session = null;
  if (sessions && sessions.length > 0) {
    session = sessions[(__VU - 1) % sessions.length];
  }

  if (!session || !session.token) {
    logError(`[Broadcast] VU ${__VU} 無法取得 session，跳過`);
    sleep(2);
    return;
  }

  // 還原 CSRF cookie（避免 cookie jar 污染）
  if (session.csrfToken) {
    const jar = http.cookieJar();
    jar.set(config.BASE_URL, "csrf_token", session.csrfToken, { path: "/" });
  }

  // ── 決定角色 ─────────────────────────────────────────
  const totalVUs = parseInt(__ENV.K6_VUS || "50", 10);
  const senderCount = Math.max(1, Math.floor(totalVUs * SENDER_RATIO));
  const isSender = __VU <= senderCount;

  logInfo(
    `[Broadcast] VU ${__VU} 角色: ${isSender ? "SENDER" : "LISTENER"} | 房間: ${broadcastChannelId}`,
  );

  const fullUrl = `${config.WS_URL}?token=${session.token}`;

  ws.connect(fullUrl, {}, function (socket) {
    let roomJoined = false;
    let sessionEnded = false; // 防止 soak timeout 和 error handler 重複關閉
    let msgSeq = 0;

    // ── 訊息監聽 ─────────────────────────────────────
    socket.on("message", function (raw) {
      let msg;
      try {
        msg = JSON.parse(raw);
      } catch (_) {
        return;
      }

      switch (msg.action) {
        case "room_joined":
          if (roomJoined) break; // 防止重複處理
          roomJoined = true;
          wsRoomJoinSuccess.add(1);
          check(null, { "broadcast: room joined successfully": () => true });
          logInfo(`[Broadcast] VU ${__VU} 加入房間成功`);

          // ── 掛網主計時器（到期後離開房間並關閉連線）────
          socket.setTimeout(function () {
            if (sessionEnded) return;
            sessionEnded = true;
            socket.send(
              JSON.stringify({
                action: "leave_room",
                data: { room_id: broadcastChannelId, room_type: "channel" },
              }),
            );
            socket.close();
          }, SOAK_DURATION_MS);

          if (isSender) {
            // Sender：定時發送訊息
            socket.setInterval(function () {
              if (sessionEnded) return;
              msgSeq++;
              const content = `BroadcastTest seq=${msgSeq}@${Date.now()} vu=${__VU}`;
              socket.send(
                JSON.stringify({
                  action: "send_message",
                  data: {
                    room_id: broadcastChannelId,
                    room_type: "channel",
                    content: content,
                  },
                }),
              );
              wsBroadcastSent.add(1);
            }, SEND_INTERVAL_MS);
          } else {
            // Listener：每 30 秒發一次 ping 保持連線
            socket.setInterval(function () {
              if (sessionEnded) return;
              socket.send(JSON.stringify({ action: "ping", data: {} }));
            }, 30000);
          }
          break;

        case "new_message":
          wsBroadcastReceived.add(1);
          // 若訊息內含 seq（由 sender 帶入），計算廣播延遲
          if (msg.data && msg.data.content) {
            const match = msg.data.content.match(/seq=(\d+)@(\d+)/);
            if (match) {
              const sentAt = parseInt(match[2], 10);
              const latency = Date.now() - sentAt;
              wsBroadcastLatency.add(latency);
            }
          }
          break;

        case "message_sent":
          // Sender 收到自己訊息的確認
          break;

        case "error":
          if (!roomJoined && !sessionEnded) {
            sessionEnded = true;
            wsRoomJoinFailed.add(1);
            check(null, { "broadcast: room joined successfully": () => false });
            logError(`[Broadcast] WS error: ${msg.data?.message}`);
            socket.close();
          }
          break;
      }
    });

    socket.on("error", function (e) {
      logError(`[Broadcast] VU ${__VU} socket error: ${e.error()}`);
      if (!roomJoined && !sessionEnded) {
        sessionEnded = true;
        wsRoomJoinFailed.add(1);
        check(null, { "broadcast: room joined successfully": () => false });
      }
    });

    // ── 加入房間 ─────────────────────────────────────
    socket.send(
      JSON.stringify({
        action: "join_room",
        data: { room_id: broadcastChannelId, room_type: "channel" },
      }),
    );

    // join 逾時計時器：若在 ROOM_JOIN_TIMEOUT_MS 內未收到 room_joined，關閉連線
    socket.setTimeout(function () {
      if (!roomJoined && !sessionEnded) {
        sessionEnded = true;
        wsRoomJoinFailed.add(1);
        check(null, { "broadcast: room joined successfully": () => false });
        logError(`[Broadcast] VU ${__VU} 加入房間逾時，放棄`);
        socket.close();
      }
    }, ROOM_JOIN_TIMEOUT_MS);
  });
}
