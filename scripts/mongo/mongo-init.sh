#!/usr/bin/env bash
set -euo pipefail

ENV_NAME="${1:-development}"
if [[ "${ENV_NAME}" == "production" ]]; then
  ENV_FILE=".env"
else
  ENV_FILE=".env.${ENV_NAME}"
fi
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if [[ ! -f "${PROJECT_ROOT}/${ENV_FILE}" ]]; then
  echo "❌ 找不到環境檔案: ${ENV_FILE}"
  exit 1
fi

set -a
source "${PROJECT_ROOT}/${ENV_FILE}"
set +a

export INIT_ENV="${ENV_NAME}"

# 管理者帳密優先使用指令參數，否則從 env 讀取
ADMIN_USER="${ADMIN_USERNAME:-${MONGO_INITDB_ROOT_USERNAME:-root}}"
ADMIN_PASS="${ADMIN_PASSWORD:-${MONGO_INITDB_ROOT_PASSWORD:-}}"

# 其他配置都從 env 讀取
MONGO_HOST="${MONGO_URI:-localhost:27017}"
MONGO_DB="${MONGO_DB_NAME:-chat_app}"

echo "[mongo-init] 環境: ${ENV_NAME}"
echo "[mongo-init] 目標: ${MONGO_HOST}/${MONGO_DB}"

# 判斷是否為完整 URI
if [[ "${MONGO_HOST}" =~ ^mongodb(\+srv)?:// ]]; then
  mongosh "${MONGO_HOST}" --file "${PROJECT_ROOT}/scripts/mongo-init.js"
else
  # Host:Port 模式，使用管理者帳密
  if [[ -n "${ADMIN_PASS}" ]]; then
    echo "[mongo-init] 使用管理者帳號: ${ADMIN_USER}"
    mongosh "mongodb://${MONGO_HOST}/${MONGO_DB}" \
      --username "${ADMIN_USER}" \
      --password "${ADMIN_PASS}" \
      --authenticationDatabase admin \
      --file "${PROJECT_ROOT}/scripts/mongo-init.js"
  else
    echo "[mongo-init] 使用匿名連線"
    mongosh "mongodb://${MONGO_HOST}/${MONGO_DB}" \
      --file "${PROJECT_ROOT}/scripts/mongo-init.js"
  fi
fi

echo "✅ Mongo 初始化完成"
