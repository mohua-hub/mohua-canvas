package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/tigerowo/infinite-canvas/config"
	"github.com/tigerowo/infinite-canvas/handler"
	"github.com/tigerowo/infinite-canvas/router"
	"github.com/tigerowo/infinite-canvas/service"
)

func main() {
	desktop := os.Getenv("MOHUA_DESKTOP") == "1"
	if desktop {
		// 桌面进程持有 stdin 写端，退出后子进程随管道关闭结束。
		go func() {
			_, _ = io.Copy(io.Discard, os.Stdin)
			os.Exit(0)
		}()
	}
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := service.EnsureDefaultAdmin(); err != nil {
		log.Fatal(err)
	}
	if err := service.EnsureDefaultAgentSkills(); err != nil {
		log.Fatal(err)
	}
	service.StartPromptSyncScheduler()
	service.StartCanvasProjectCleanupScheduler()
	handler.StartVideoTaskPoller()
	engine := router.New()
	if desktop {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("MOHUA_API_READY=http://" + listener.Addr().String())
		log.Fatal(engine.RunListener(listener))
	}
	log.Fatal(engine.Run(":" + config.Cfg.Port))
}
