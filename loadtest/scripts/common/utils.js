/**
 * 通用工具函數
 */
import { sleep } from 'k6';

/**
 * 從陣列中隨機選擇一個元素
 * @param {Array} arr - 來源陣列
 * @returns {*} 陣列中的隨機元素
 */
export function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

/**
 * 生成指定長度的隨機字串
 * @param {number} length - 字串長度
 * @returns {string} 隨機字串
 */
export function randomString(length) {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * 在指定範圍內隨機延遲
 * @param {number} min - 最小延遲秒數
 * @param {number} max - 最大延遲秒數
 */
export function randomSleep(min = 1, max = 3) {
  sleep(Math.random() * (max - min) + min);
}
