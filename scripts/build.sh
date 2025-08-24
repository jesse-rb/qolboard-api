export PATH=$PATH:/usr/local/go/bin
cp /.env /app/.env
cd /app
go mod init qolboard-api
go mod tidy
go build -o /app/server /app/main.go
