package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"swallow/app"
	"swallow/config"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uopensail/ulib/prome"
	"github.com/uopensail/ulib/zlog"
	"google.golang.org/grpc"
)

func run(configFilePath string, logDir string) *app.App {
	config.AppConf.Init(configFilePath)
	zlog.InitLogger(config.AppConf.ProjectName, config.AppConf.Debug, logDir)

	app := app.NewApp()
	runGRPC(app.GRPCAPIRegister)
	runHTTPServe(app.RegisterGinRouter)
	return app
}

func runGRPC(registerFunc func(server *grpc.Server)) {
	go func() {
		grpcServer := grpc.NewServer()
		registerFunc(grpcServer)
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d",
			config.AppConf.ServerConfig.GRPCPort))
		if err != nil {
			zlog.SLOG.Error(err)
			panic(err)
		}
		fmt.Printf("[info] start grpc server listening %v\n", listener.Addr())
		err = grpcServer.Serve(listener)
		if err != nil {
			zlog.SLOG.Error(err)
			panic(err)
		}
	}()
}

func runHTTPServe(registerFunc func(*gin.Engine)) {
	go func() {
		ginEngine := gin.New()
		ginEngine.Use(gin.Recovery())
		conf := config.AppConf.ServerConfig.HttpServerConfig

		registerFunc(ginEngine)

		err := ginEngine.Run(fmt.Sprintf(":%d", conf.HTTPPort))
		if err != nil {
			zlog.SLOG.Error(err)
			panic(err)
		}
	}()
}

func runPProf(port int) {
	if port > 0 {
		go func() {
			fmt.Println(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil))
		}()
	}
}

func runProme(projectName string, port int) *prome.Exporter {
	//prome的打点
	promeExport := prome.NewExporter(projectName)
	go func() {
		err := promeExport.Start(port)
		if err != nil {
			panic(err)
		}
	}()

	return promeExport
}

func main() {
	configFilePath := flag.String("config", "conf/config.toml", "启动命令请设置配置文件目录")
	logDir := flag.String("log", "./logs", "启动命令请设置seelog.xml")
	flag.Parse()

	app := run(*configFilePath, *logDir)

	if len(config.AppConf.ProjectName) <= 0 {
		panic("config.ProjectName NULL")
	}

	runPProf(config.AppConf.PProfPort)
	promeExport := runProme(config.AppConf.ProjectName, config.AppConf.PromePort)
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " app running....")

	<-signalChanel

	app.Close()
	promeExport.Close()

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " app exit....")
}
