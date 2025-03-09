package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"tomestobot/internal/bot"
	bxwrapper "tomestobot/internal/bx"
)

func main() {
	err := mainRun()
	if err != nil {
		log.Printf("Main finished with error: %s", err.Error())
	}
}

func mainRun() error {
	bot, err := bot.New(os.Getenv("TG_TOKEN"))
	if err != nil {
		return fmt.Errorf("new bot: %w", err)
	}
	return bot.Start()
}

func bxTest() error {
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

	if len(deals) == 0 {
		return fmt.Errorf("no deals")
	}

	// Get tasks for deal
	tasks, err := u.ListDealTasks(deals[0].Id)
	if err != nil {
		return fmt.Errorf("list deal tasks: %w", err)
	}
	log.Print(tasks)

	// if len(tasks) == 0 {
	// 	return fmt.Errorf("no tasks")
	// }

	// Complete task
	// if err := u.CompleteTask(tasks[0].Id); err != nil {
	// 	return fmt.Errorf("complete task: %w", err)
	// }

	// Leave a comment
	// if err := u.AddCommentToDeal(deals[0].Id, "Another test comment and now via golang"); err != nil {
	// 	return fmt.Errorf("add comment to deal: %w", err)
	// }

	return nil
}
