package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"tomestobot/internal/bx"
)

func main() {
	err := mainRun()
	if err != nil {
		log.Printf("Main finished with error: %s", err.Error())
	}
}

func mainRun() error {
	// Creating bitrix wrapper
	userId, err := strconv.Atoi(os.Getenv("BX_USER_ID"))
	if err != nil {
		return fmt.Errorf("invalid user id env variable: %w", err)
	}
	bx, err := bx.New(os.Getenv("BX_URL"), userId, os.Getenv("BX_TOKEN"))
	if err != nil {
		return fmt.Errorf("bx creation: %w", err)
	}

	_ = bx
	return nil
}
