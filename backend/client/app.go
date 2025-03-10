package client

import (
	"bufio"
	"context"
	"doki-byte/FreeProxy/backend/config"
	"doki-byte/FreeProxy/backend/info"
	"doki-byte/FreeProxy/backend/request"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"strings"
	"time"
)

type App struct {
	ctx      context.Context
	stopChan chan struct{}

	rpm     *request.ProxyManager
	metrics *info.ProxyMetrics
	config  *config.Config
}

func NewApp() *App {
	stopChan := make(chan struct{})
	return &App{
		stopChan: stopChan,
		config:   config.GetConfig(),
		rpm: request.NewProxyManager([]request.ProxyFetcher{
			//&request.Free89,
			//&request.FreeHappy,
			//&request.FreeQiYun,
			&request.HunterConfig{},
			&request.QuakeConfig{},
			&request.FofaConfig{},
		}),
		metrics: info.NewProxyMetrics(stopChan),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go a.metrics.StartMonitoring(1 * time.Second)
	go a.metricsPusher()
}

func (a *App) Shutdown(ctx context.Context) {
	a.metrics.StopMonitoring()
	close(a.stopChan)
}

func (a *App) GetProfile() config.Config {
	return a.config.GetProfile()
}

func (a *App) FetchProxies() Response {
	proxies, err := a.rpm.FetchAll()
	if err != nil {
		return a.errorResponse(err)
	}

	table, err := a.rpm.RenderTable()
	if err != nil {
		return a.errorResponse(err)
	}

	a.config.SetAllProxies(proxies)
	return Response{
		Code:    200,
		Message: "抓取成功",
		Data:    string(table),
	}
}

func (a *App) ChooseFile() config.Config {
	a.config.Code = 200

	//// 获取配置文件路径
	//optSys := runtime2.GOOS
	//proxy_success_path := ""
	//if optSys == "windows" {
	//	proxy_success_path = config.GetCurrentAbPathByExecutable() + "\\proxy_success.txt"
	//} else {
	//	proxy_success_path = config.GetCurrentAbPathByExecutable() + "/proxy_success.txt"
	//}
	//if _, err := os.Stat(proxy_success_path); err == nil {
	//	a.config.FilePath = proxy_success_path
	//} else {
	//}

	a.config.FilePath, _ = runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:           "请选择配置文件",
		ShowHiddenFiles: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "配置文件", Pattern: "*.txt"},
		},
	})
	if a.config.FilePath == "" {
		a.config.Code = 400
		a.config.Error = "未选择配置文件"
		return a.config.GetProfile()
	}

	f, errOpen := os.Open(a.config.FilePath)
	if errOpen != nil {
		a.config.Code = 400
		a.config.Error = errOpen.Error()
		return a.config.GetProfile()
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.Size() == 0 {
		runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[WARN] 配置文件 %s 是空的", a.config.FilePath))
		a.config.Code = 400
		a.config.Error = "配置文件为空"
		return a.config.GetProfile()
	}

	// 按行读取代理数据
	var lists []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 忽略空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lists = append(lists, line)
	}

	if err := scanner.Err(); err != nil {
		a.config.Code = 400
		a.config.Error = err.Error()
		return a.config.GetProfile()
	}

	a.config.SetAllProxies(lists)
	runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[INF] 配置文件 %s 读取成功，共 %d 条数据", a.config.FilePath, len(lists)))

	rsp := a.CheckDatasets()
	if rsp.Code != 200 {
		a.Error("测试有误： %s\n", rsp.Message)
		a.config.Code = 400
		a.config.Error = rsp.Message
		return a.config.GetProfile()
	}

	return a.config.GetProfile()
}

func (a *App) UseFetchedDatasets() Response {
	a.StopListening()
	a.stopTask()
	runtime.EventsEmit(a.ctx, "log_update", "[INF] 使用抓取的代理")
	return a.CheckDatasets()
}

func (a *App) metricsPusher() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		runtime.EventsEmit(a.ctx, "metrics_update", a.metrics.GetMetrics())
	}
}

func (a *App) SaveConfig(data config.Config) string {
	a.config = &data
	a.StopListening()
	a.stopTask()
	err := a.config.SaveConfig()
	if err != nil {
		return "保存失败: " + err.Error()
	}
	return "保存成功"
}
