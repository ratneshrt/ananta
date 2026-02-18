#!/bin/bash
set -e

APP_NAME=$1
REPO=$2
RUNTIME=$3
SUBDOMAIN=$4
PORT=$5

BASE_DIR="/srv/apps/$APP_NAME"
VENV="$BASE_DIR/venv"
ROUTES_FILE="/srv/router/routes.txt"
ROUTER_RELOAD="http://localhost:9001/__reload"

mkdir -p /srv/apps

if [ ! -d "$BASE_DIR/.git" ]; then
  git clone "$REPO" "$BASE_DIR"
else
  cd "$BASE_DIR"
  git pull
fi

cd "$BASE_DIR"

pm2 delete "$APP_NAME" >/dev/null 2>&1 || true

if [ "$RUNTIME" = "node" ]; then
  npm install
  PORT=$PORT pm2 start index.js --name "$APP_NAME"

elif [ "$RUNTIME" = "go" ]; then
  go build -o "$APP_NAME"
  PORT=$PORT pm2 start "./$APP_NAME" --name "$APP_NAME"

elif [ "$RUNTIME" = "python" ]; then
  if [ ! -d "$VENV" ]; then
    python3 -m venv "$VENV"
  fi

  source "$VENV/bin/activate"

  if [ -f requirements.txt ]; then
    pip install -r requirements.txt
  fi

  PORT=$PORT pm2 start app.py \
    --interpreter "$VENV/bin/python" \
    --name "$APP_NAME"

else
  echo "Unsupported runtime: $RUNTIME"
  exit 1
fi

grep -v "^$SUBDOMAIN\.zenops\.in=" "$ROUTES_FILE" > /tmp/routes.tmp || true
echo "$SUBDOMAIN.zenops.in=$PORT" >> /tmp/routes.tmp
mv /tmp/routes.tmp "$ROUTES_FILE"

curl -s "$ROUTER_RELOAD" > /dev/null

echo "https://$SUBDOMAIN.zenops.in"
