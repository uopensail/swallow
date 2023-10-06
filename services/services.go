package services

import (
	"context"
	"errors"
	"net/http"
	"swallow/config"
	"swallow/warehouse"

	"github.com/gin-gonic/gin"
	"github.com/uopensail/swallow-idl/proto/swallowapi"
	"github.com/uopensail/ulib/prome"
	"github.com/uopensail/ulib/utils"
	etcdclient "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var __GITHASH__ = ""

type Services struct {
	w *warehouse.Warehouse
	swallowapi.UnimplementedSwallowServiceServer
}

func NewServices() *Services {
	return &Services{
		w: warehouse.NewWarehouse(config.AppConfigInstance.WorkDir, config.AppConfigInstance.PrimaryKey),
	}
}
func (srv *Services) Init(configFolder string, etcdName string, etcdCli *etcdclient.Client, reg utils.Register) {

}
func (srv *Services) RegisterGrpc(grpcS *grpc.Server) {
	swallowapi.RegisterSwallowServiceServer(grpcS, srv)

}

func (srv *Services) RegisterGinRouter(ginEngine *gin.Engine) {
	ginEngine.POST("/trace", srv.TraceEchoHandler)
	ginEngine.POST("/log", srv.LogEchoHandler)
}
func (srv *Services) Close() {
	srv.w.Close()
}

func (srv *Services) Trace(ctx context.Context, in *swallowapi.Request) (*swallowapi.Response, error) {
	stat := prome.NewStat("App.Trace")
	defer stat.End()
	if in == nil || len(in.Data) <= 0 {
		return nil, errors.New("input empy")
	}
	srv.w.Trace(in.Data)
	return &swallowapi.Response{
		Code: 200,
	}, nil
}

func (srv *Services) Log(ctx context.Context, in *swallowapi.Request) (*swallowapi.Response, error) {
	stat := prome.NewStat("App.Log")
	defer stat.End()
	srv.w.Log(in.Data)
	return &swallowapi.Response{
		Code: 200,
	}, nil
}

func (srv *Services) TraceEchoHandler(gCtx *gin.Context) {
	stat := prome.NewStat("App.TraceEchoHandler")
	defer stat.End()
	request := &swallowapi.Request{}
	if err := gCtx.ShouldBind(&request); err != nil {
		stat.MarkErr()
		gCtx.JSON(http.StatusInternalServerError, swallowapi.Response{
			Code: -1,
		})
		return
	}

	response, err := srv.Trace(context.Background(), request)
	if err != nil {
		stat.MarkErr()
		return
	}
	gCtx.JSON(http.StatusOK, response)
	return
}

func (srv *Services) LogEchoHandler(gCtx *gin.Context) {
	stat := prome.NewStat("App.LogEchoHandler")
	defer stat.End()
	request := &swallowapi.Request{}
	if err := gCtx.ShouldBind(&request); err != nil {
		stat.MarkErr()
		gCtx.JSON(http.StatusInternalServerError, swallowapi.Response{
			Code: -1,
		})
		return
	}

	response, err := srv.Log(context.Background(), request)
	if err != nil {
		stat.MarkErr()
		return
	}
	gCtx.JSON(http.StatusOK, response)
	return
}

func (srv *Services) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
	}, nil
}

func (srv *Services) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
	return nil
}
