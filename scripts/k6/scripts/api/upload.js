/**
 * API 測試 - 檔案上傳 (Upload)
 */
import { group, check } from "k6";
import http from "k6/http";
import { applyCsrf } from "../common/csrf.js";
import {
  logHttpResponse,
  logGroupStart,
  logGroupEnd,
  logInfo,
} from "../common/logger.js";

export default function (baseUrl, session) {
  if (!session) return;

  const groupStartTime = Date.now();
  logGroupStart("File Upload APIs");

  group("API - File Upload", function () {
    group("Upload General File", function () {
      // 根據 API.md: POST /upload/file
      logInfo("上傳一般檔案: test-document.txt (1KB)");
      const url = `${baseUrl}/upload/file`;

      // 創建一個模擬檔案
      const data = {
        file: http.file(
          new ArrayBuffer(1024),
          "test-document.txt",
          "text/plain"
        ),
      };

      const headers = {
        ...applyCsrf(url, {}),
        Authorization: session.headers["Authorization"],
        "Origin": "http://localhost:3000",
      };
      // 注意：multipart 請求不需要手動設定 Content-Type

      const res = http.post(url, data, { headers });
      let body;
      try {
        body = res.json();
      } catch (e) {
        body = null;
      }

      logHttpResponse("POST /upload/file", res, { expectedStatus: 200 });

      check(res, {
        "Upload File: status is 200": (r) => r.status === 200,
        "Upload File: response has success status": () =>
          body && body.status === "success",
        "Upload File: has file_url": () =>
          body && body.data && body.data.file_url !== undefined,
      });

      if (res.status === 200) {
        logInfo("✅ 一般檔案上傳成功");
      }
    });

    group("Upload Avatar", function () {
      // 根據 API.md: POST /upload/avatar
      logInfo("上傳頭像檔案: avatar.jpg (2KB)");
      const url = `${baseUrl}/upload/avatar`;

      // 創建一個模擬頭像檔案
      const data = {
        avatar: http.file(new ArrayBuffer(2048), "avatar.jpg", "image/jpeg"),
      };

      const headers = {
        ...applyCsrf(url, {}),
        Authorization: session.headers["Authorization"],
        "Origin": "http://localhost:3000",
      };

      const res = http.post(url, data, { headers });
      let body;
      try {
        body = res.json();
      } catch (e) {
        body = null;
      }

      logHttpResponse("POST /upload/avatar", res, { expectedStatus: 200 });

      check(res, {
        "Upload Avatar: status is 200": (r) => r.status === 200,
        "Upload Avatar: response has success status": () =>
          body && body.status === "success",
        "Upload Avatar: has file_url": () =>
          body && body.data && body.data.file_url !== undefined,
      });

      if (res.status === 200) {
        logInfo("✅ 頭像檔案上傳成功");
      }
    });

    group("Upload Document", function () {
      // 根據 API.md: POST /upload/document
      logInfo("上傳文件檔案: report.pdf (5KB)");
      const url = `${baseUrl}/upload/document`;

      // 創建一個模擬文件
      const data = {
        document: http.file(
          new ArrayBuffer(5120),
          "report.pdf",
          "application/pdf"
        ),
      };

      const headers = {
        ...applyCsrf(url, {}),
        Authorization: session.headers["Authorization"],
        "Origin": "http://localhost:3000",
      };

      const res = http.post(url, data, { headers });
      let body;
      try {
        body = res.json();
      } catch (e) {
        body = null;
      }

      logHttpResponse("POST /upload/document", res, { expectedStatus: 200 });

      check(res, {
        "Upload Document: status is 200": (r) => r.status === 200,
        "Upload Document: response has success status": () =>
          body && body.status === "success",
        "Upload Document: has file_url": () =>
          body && body.data && body.data.file_url !== undefined,
      });

      if (res.status === 200) {
        logInfo("✅ 文件檔案上傳成功");
      }
    });

    group("Get Files List", function () {
      // 根據 API.md: GET /files
      logInfo("取得檔案列表");
      const headers = {
        ...session.headers,
        ...applyCsrf(`${baseUrl}/files`, session.headers),
        "Origin": "http://localhost:3000",
      };
      const res = http.get(`${baseUrl}/files`, { headers });

      logHttpResponse("GET /files", res, { expectedStatus: 200 });

      check(res, {
        "Get Files: status is 200": (r) => r.status === 200,
        "Get Files: response has success status": (r) =>
          r.json("status") === "success",
      });

      if (res.status === 200) {
        logInfo("✅ 檔案列表取得成功");
      }
    });
  });

  logGroupEnd("File Upload APIs", groupStartTime);
}
