default: run

build:
	@go build -o ./out/tomestobot.exe ./cmd/

run: build
	@./out/tomestobot.exe
