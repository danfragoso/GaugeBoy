.PHONY: go
go:
	CC=arm-linux-gnueabihf-gcc CGO_ENABLED=1 GOARCH=arm GOOS=linux go build -o App/GaugeBoy/GaugeBoy src/main.go