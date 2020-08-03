env GOOS=linux GOARCH=amd64 go build -o nest-linux-amd64 main.go
mv nest-linux-amd64 bin/

env GOOS=darwin GOARCH=amd64 go build -o nest-darwin-amd64 main.go
mv nest-darwin-amd64 bin/
