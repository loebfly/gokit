package app

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/loebfly/ezgin/internal/config"
	"github.com/loebfly/ezgin/internal/logs"
	"github.com/loebfly/ezgin/internal/nacos"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	servers = make([]*http.Server, 0)
)

// Start 启动服务
func (receiver enter) Start(ymlPath string, ginEngine *gin.Engine) {

	receiver.initEZGin(ymlPath, ginEngine)
	ez := config.EZGin()

	logs.Enter.CInfo("APP", "|-----------------------------------|")
	logs.Enter.CInfo("APP", "| 服务名: {}", ez.App.Name)
	logs.Enter.CInfo("APP", "| 版本号: {}", ez.App.Version)
	logs.Enter.CInfo("APP", "|-----------------------------------|")
	if ez.App.Port > 0 {
		logs.Enter.CInfo("APP", "| HTTP端口: {}", ez.App.Port)
	}
	if ez.App.PortSsl > 0 {
		logs.Enter.CInfo("APP", "| HTTPS端口: {}", ez.App.PortSsl)
	}
	logs.Enter.CInfo("APP", "|-----------------------------------|")
}

// ShutdownWhenExitSignal 服务异常退出时 优雅关闭服务
func (receiver enter) ShutdownWhenExitSignal(will func(os.Signal), did func(context.Context)) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signalChan
	logs.Enter.CError("APP", "收到退出信号:{}", sig.String())
	nacos.Enter.UnregisterIfNeed()

	if will != nil {
		will(sig)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	closeAllSuccess := true
	if len(servers) == 0 {
		closeAllSuccess = false
	}
	for _, server := range servers {
		if server != nil {
			if err := server.Shutdown(ctx); err != nil {
				closeAllSuccess = false
				logs.Enter.CError("APP", "关闭服务失败:{}", err.Error())
				return
			}
		}
	}
	if closeAllSuccess {
		logs.Enter.CError("APP", "服务已关闭")
	}

	if did != nil {
		did(ctx)
	}
}