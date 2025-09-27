#!/bin/bash
#又到了我最喜欢的编写自动化sh脚本的环节！
set -e

# 检测是否在 Docker 环境
# 容器启动时，Docker 会在根目录创建一个隐藏文件,或者在 cgroup 信息中包含 docker 或 kubepods 字样
# 但在极个别情况下不会生成/.dockerenv 文件，所以加上 cgroup 的判断
au_in_docker() {
  if [ -f "/.dockerenv" ] || grep -qE '(docker|kubepods)' /proc/1/cgroup 2>/dev/null; then
    return 0
  fi
  return 1
}

if au_in_docker; then
  DATA_DIR="/data"
else
  DATA_DIR="./data"
  mkdir -p "$DATA_DIR"
fi

SECRET_FILE="$DATA_DIR/jwt_secret"

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

# 在宿主机直接执行时，可能不在 /app，需要修正路径
if [ -x "/app/confession-wall" ]; then
  exec /app/confession-wall
elif [ -x "./confession-wall" ]; then
  exec ./confession-wall
else
  echo "[ERROR] confession-wall binary not found!"
  exit 1
fi
