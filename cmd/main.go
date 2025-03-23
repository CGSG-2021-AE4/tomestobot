package main

import (
	"fmt"
	"log/slog"
	"strings"

	"os"
	"strconv"

	"github.com/CGSG-2021-AE4/tomestobot/api"
	"github.com/CGSG-2021-AE4/tomestobot/internal/bot"
	"github.com/CGSG-2021-AE4/tomestobot/internal/bx"
	"github.com/CGSG-2021-AE4/tomestobot/pkg/log"
)

func main() {
	api.SetupGlobalFlags() // Setups some env flags

	if err := mainRun(); err != nil {
		slog.Error("Main finished with error", "err", err.Error())
	}
}

func mainRun() error {
	// Parse env variable
	userId, err := strconv.Atoi(os.Getenv("BX_USER_ID"))
	if err != nil {
		return fmt.Errorf("invalid user id env variable: %w", err)
	}

	// Setup logger
	logsLevel := slog.LevelInfo
	if api.EnableDebugLogs {
		logsLevel = slog.LevelDebug
	}
	defferedOutput := log.NewDefferedOutput()
	output := log.NewMultiOutput(log.NewConsoleLogOutput(true), defferedOutput)
	logger := slog.New(log.NewHandler(output, logsLevel))

	// Create bx wrapper

	bxDescr := bx.BxDescriptor{
		BxDomain: os.Getenv("BX_DOMAIN"),
		BxUserId: userId,
		BxHook:   os.Getenv("BX_HOOK"),
	}
	bx, err := bx.New(logger.WithGroup("BX"), bxDescr)
	if err != nil {
		return fmt.Errorf("bx creation")
	}

	// Create bot

	botDescr := bot.BotDescriptor{
		TgBotToken:     os.Getenv("TG_TOKEN"),
		Bx:             bx,
		AdminWhitelist: strings.Split(os.Getenv("ADMIN_WHITELIST"), " "),
	}
	bot, err := bot.New(logger.WithGroup("TG"), botDescr)
	if err != nil {
		return fmt.Errorf("new bot: %w", err)
	}

	// Setup tg logging
	defferedOutput.Output = bot.GetLogsOutput()
	return bot.Start()
}

func bxTest() error {
	// Creating bitrix wrapper
	//	userId, err := strconv.Atoi(os.Getenv("BX_USER_ID"))
	//	if err != nil {
	//		return fmt.Errorf("invalid user id env variable: %w", err)
	//	}
	//	bx, err := bxwrapper.New(os.Getenv("BX_URL"), userId, os.Getenv("BX_TOKEN"))
	//	if err != nil {
	//		return fmt.Errorf("bx creation: %w", err)
	//	}
	//
	//	// Auth user
	//	u, err := bx.AuthUserByPhone(os.Getenv("TEST_PHONE"))
	//	if err != nil {
	//		return fmt.Errorf("auth user by phone: %w", err)
	//	}
	//	log.Print(u.GetId())
	//
	//	// Get deals
	//	deals, err := u.ListDeals()
	//	if err != nil {
	//		return fmt.Errorf("list deals: %w", err)
	//	}
	//	log.Print(deals)
	//
	//	if len(deals) == 0 {
	//		return fmt.Errorf("no deals")
	//	}
	//
	//	// Get tasks for deal
	//	tasks, err := u.ListDealTasks(deals[0].Id)
	//	if err != nil {
	//		return fmt.Errorf("list deal tasks: %w", err)
	//	}
	//	log.Print(tasks)
	//
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
