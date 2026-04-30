package main

import (
	"flag"
	"fmt"

	"eth-valuation-api/internal/config"
	"eth-valuation-api/internal/handler"
	"eth-valuation-api/internal/middleware"
	"eth-valuation-api/internal/scheduler"
	"eth-valuation-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// Register global middlewares: CORS first, then panic recovery.
	server.Use(middleware.CorsMiddleware)
	server.Use(middleware.RecoveryMiddleware)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// Start the background data refresh scheduler.
	sched := scheduler.NewScheduler(ctx)
	sched.Start()
	defer sched.Stop()

	fmt.Printf("Starting valuation-api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
