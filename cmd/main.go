package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	bxwrapper "tomestobot/internal/bx"
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
	bx, err := bxwrapper.New(os.Getenv("BX_URL"), userId, os.Getenv("BX_TOKEN"))
	if err != nil {
		return fmt.Errorf("bx creation: %w", err)
	}

	// Auth user
	u, err := bx.AuthUserByPhone(os.Getenv("TEST_PHONE"))
	if err != nil {
		return fmt.Errorf("auth user by phone: %w", err)
	}
	log.Print(u.GetId())

	// Get deals
	deals, err := u.ListDeals()
	if err != nil {
		return fmt.Errorf("list deals: %w", err)
	}
	log.Print(deals)

	return nil
}
