import ws from 'k6/ws';
import { check } from 'k6';
import { logInfo, logError, logWsMessage } from '../common/logger.js';

export function chatTest(baseUrl, authToken) {
  const url = baseUrl.replace('http', 'ws') + '/ws';
  const params = { headers: { 'Authorization': `Bearer ${authToken}` } };

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      logInfo('WebSocket connection established!');
      
      // 加入一個測試房間
      const joinRoomPayload = {
        type: 'join_room',
        data: { room_id: 'your_test_room_id' }
      };
      socket.send(JSON.stringify(joinRoomPayload));
      logWsMessage('Sent', joinRoomPayload);

      // 每隔幾秒發送一條訊息
      socket.setInterval(() => {
        const sendMessagePayload = {
          type: 'send_message',
          data: {
            room_id: 'your_test_room_id',
            content: `Hello from k6 user ${__VU} at ${new Date()}`
          }
        };
        socket.send(JSON.stringify(sendMessagePayload));
        logWsMessage('Sent', sendMessagePayload);
      }, 3000); // 每3秒發送一次
    });

    socket.on('message', (data) => {
      const message = JSON.parse(data);
      logWsMessage('Received', message);
      check(message, {
        'Received message has a type': (m) => m.type !== undefined,
      });
    });

    socket.on('close', () => {
      logInfo('WebSocket connection closed.');
    });

    socket.on('error', function (e) {
      logError(`WebSocket error: ${e.error()}`);
    });

    // 30秒後自動關閉連線
    socket.setTimeout(function () {
      socket.close();
    }, 30000);
  });

  check(res, { 'WebSocket handshake successful': (r) => r && r.status === 101 });
}