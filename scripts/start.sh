cd /app
mkdir -p /app/logs
nohup ./server >> /app/logs/server.log 2>&1 &
echo $! >/app/.apppid
