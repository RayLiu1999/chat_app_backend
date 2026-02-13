/**
 * 資料準備腳本：預先建立並登入測試用戶
 * 使用方式：
 * k6 run prepare_users.js --env USER_COUNT=500
 */
import * as config from './config.js';
import { getAuthenticatedSessionWithOptions } from './scripts/common/auth.js';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  const baseUrl = `${config.TEST_CONFIG.BASE_URL}${config.TEST_CONFIG.API_PREFIX}`;
  
  // 嘗試讀取預定義用戶數，若失敗則預設為 5
  let defaultCount = 5;
  try {
    const usersData = JSON.parse(open('./data/users.json'));
    defaultCount = (usersData && usersData.length) || 5;
  } catch (e) {
    // 忽略
  }
  
  const userCount = parseInt(__ENV.USER_COUNT || `${defaultCount}`, 10);

  console.log(`🚀 開始資料準備，目標用戶數: ${userCount}`);
  let successCount = 0;

  for (let index = 1; index <= userCount; index++) {
    const session = getAuthenticatedSessionWithOptions(baseUrl, {
      userIndex: index,
      registerIfMissing: true,
    });

    if (session && session.token) {
      successCount++;
    } else {
      console.error(`❌ 用戶準備失敗: index=${index}`);
    }
  }

  console.log(`✅ 資料準備完成，成功: ${successCount}/${userCount}`);

  if (successCount === 0) {
    throw new Error('資料準備失敗：沒有任何有效用戶可用');
  }
}
