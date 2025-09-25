#!/bin/bash
#又到了我最喜欢的编写自动化sh脚本的环节！
set -e

SECRET_FILE="/data/jwt_secret"

# 如果外部没有传入 JWT_SECRET，就从文件里加载或生成
if [ -z "$JWT_SECRET" ]; then
  if [ -f "$SECRET_FILE" ]; then
    # 已存在文件，直接读取
    export JWT_SECRET=$(cat "$SECRET_FILE")
    echo "[INFO] Loaded JWT_SECRET from $SECRET_FILE"
  else
    # 没有文件，生成一个新的并保存
    export JWT_SECRET=$(openssl rand -base64 32)
    echo "$JWT_SECRET" > "$SECRET_FILE"
    echo "[INFO] No JWT_SECRET provided. Generated a new one and saved to $SECRET_FILE"
  fi
else
  echo "[INFO] Using provided JWT_SECRET (not persisted)."
fi

exec /app/confession-wall
