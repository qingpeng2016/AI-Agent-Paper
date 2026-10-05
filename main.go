package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	bot "github.com/qingpeng2016/ai-token-mall/application/bot"
	"github.com/qingpeng2016/ai-token-mall/boot"
	"github.com/qingpeng2016/ai-token-mall/common/constants"
	"github.com/qingpeng2016/ai-token-mall/common/dederi/logger"
	"github.com/qingpeng2016/ai-token-mall/common/notification"
	"github.com/qingpeng2016/ai-token-mall/conf"
	"github.com/qingpeng2016/ai-token-mall/interfaces/rest"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()
	time.Local = constants.AppLocation

	container := boot.BuildContainer()

	if conf.IsRunBot() {
		var function any = func(entry *bot.Entry) error {
			go func() {
				if err := entry.Start(); err != nil {
					notification.SendErrorLog(context.Background(), "bot start failed", zap.Error(err))
				}
			}()
			return nil
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "run bot failed", zap.String("error", err.Error()))
			panic(err)
		}
	} else {
		var function any = func(router *rest.Router) {
			router.Run(conf.GetServeConf())
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "run http server failed", zap.String("error", err.Error()))
			panic(err)
		}
	}

	gracefulExit(container)
}

func gracefulExit(container *dig.Container) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.InfoZ(context.Background(), "graceful-shutdown-start")

	if conf.IsRunBot() {
		var function any = func(entry *bot.Entry) error {
			return entry.Stop()
		}
		if err := container.Invoke(function); err != nil {
			logger.ErrorZ(context.Background(), "stop bot failed", zap.String("error", err.Error()))
		}
	} else {
		var function any = func(router *rest.Router) {
			router.Close()
		}
		if err := container.Invoke(function); err != nil {
			notification.SendErrorLog(context.Background(), "stop http server failed", zap.String("error", err.Error()))
		}
	}

	logger.InfoZ(context.Background(), "graceful-shutdown-completed")
}
