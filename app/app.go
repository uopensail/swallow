package app

import (
	"context"
	"net/http"
	"swallow/api"
	"swallow/config"
	"swallow/warehouse"

	"github.com/gin-gonic/gin"
	"github.com/uopensail/ulib/prome"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var __GITHASH__ = ""

type App struct {
	w *warehouse.Warehouse
}

func NewApp() *App {
	return &App{
		w: warehouse.NewWarehouse(config.AppConf.WorkDir, config.AppConf.PrimaryKey),
	}
}

func (app *App) GRPCAPIRegister(s *grpc.Server) {
	api.RegisterSwallowServiceServer(s, app)
}

func (app *App) Close() {
	app.w.Close()
}

func (app *App) RegisterGinRouter(ginEngine *gin.Engine) {
	ginEngine.POST("/put", app.PutEchoHandler)
	ginEngine.GET("/", app.PingEchoHandler)
	ginEngine.GET("/version", app.VersionEchoHandler)
}

func (app *App) Put(ctx context.Context, in *api.Request) (*api.Response, error) {
	stat := prome.NewStat("App.Put")
	defer stat.End()
	app.w.Put(in)
	return &api.Response{
		Code: 200,
	}, nil
}

func (app *App) PutEchoHandler(gCtx *gin.Context) {
	stat := prome.NewStat("App.PutEchoHandler")
	defer stat.End()
	request := &api.Request{}
	if err := gCtx.ShouldBind(&request); err != nil {
		stat.MarkErr()
		gCtx.JSON(http.StatusInternalServerError, api.Response{
			Code: -1,
		})
		return
	}

	response, err := app.Put(context.Background(), request)
	if err != nil {
		stat.MarkErr()
		return
	}
	gCtx.JSON(http.StatusOK, response)
	return
}

func (app *App) PingEchoHandler(gCtx *gin.Context) {
	gCtx.String(http.StatusOK, "OK")
}

func (app *App) VersionEchoHandler(gCtx *gin.Context) {
	gCtx.String(http.StatusOK, __GITHASH__)
}

func (app *App) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	}, nil
}

func (app *App) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
	return nil
}
