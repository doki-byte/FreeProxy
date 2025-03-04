package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var listener net.Listener     // 全局监听器
var cancel context.CancelFunc // 用于取消监听的上下文

// 启动监听
func (a *App) startListening() Response {
	// 检查缓存代理数
	if len(a.config.LiveProxyLists) == 0 {
		a.Error("缓存代理数为0，任务取消。")
		runtime.EventsEmit(a.ctx, "log_update", "[ERR] 缓存代理数为0。")
		runtime.EventsEmit(a.ctx, "log_update", "=========================== 任务取消 ============================")
		return a.errorResponse("缓存代理数为0，任务取消。")
	}

	// 检查监听器是否已存在
	if a.config.GetStatus() == 2 {
		// 取消监听
		cancel()
		listener.Close() // 关闭监听器
		//a.Error("监听服务已经在运行，请先停止任务。")
		//runtime.EventsEmit(a.ctx, "log_update", "[ERR] 监听服务已经在运行。")
		//runtime.EventsEmit(a.ctx, "log_update", "=========================== 任务取消 ============================")
		//return a.errorResponse("监听服务已经在运行，请先停止任务。")
	}

	// 创建监听器
	var err error
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background()) // 创建带取消功能的上下文

	listener, err = net.Listen("tcp", a.config.SocksAddress)
	if err != nil {
		a.Error("Error: %s\n", err.Error())
		runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[ERR] 监听失败 %s ", err.Error()))
		runtime.EventsEmit(a.ctx, "log_update", "=========================== 任务取消 ============================")
		return a.errorResponse(err.Error())
	}
	defer listener.Close()

	runtime.EventsEmit(a.ctx, "log_update", "=========================== 开始监听 ============================")
	runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[INF] 开始监听 socks5://%s -- 挂上代理以使用", a.config.SocksAddress))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, a.config.CoroutineCount)

	// 监听连接
	for {
		select {
		case <-ctx.Done(): // 如果上下文被取消，退出监听
			runtime.EventsEmit(a.ctx, "log_update", "[INF] 监听已停止")
			wg.Wait() // 等待所有连接处理完成
			return a.successResponse("监听已成功停止", nil)
		default:
			conn, err := listener.Accept()
			if err != nil {
				runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[ERR] 接受连接失败 %s ", err.Error()))
				continue
			}

			semaphore <- struct{}{}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				defer func() { <-semaphore }()
				a.handleConnection(conn)
			}(conn)
		}
	}
}

// 停止监听
func (a *App) StopListening() Response {
	if listener == nil {
		a.Error("监听服务未启动")
		runtime.EventsEmit(a.ctx, "log_update", "[ERR] 监听服务未启动")
		return a.errorResponse("监听服务未启动")
	}

	// 取消监听
	cancel()
	listener.Close() // 关闭监听器

	runtime.EventsEmit(a.ctx, "log_update", "[INF] 监听已停止")
	return a.successResponse("监听已成功停止", nil)
}

// 处理连接
func (a *App) handleConnection(conn net.Conn) {
	defer conn.Close()

	if len(a.config.LiveProxyLists) == 0 {
		runtime.EventsEmit(a.ctx, "log_update", "[ERR] 没有可用代理")
		return
	}

	current := a.config.LiveProxyLists[rand.Intn(len(a.config.LiveProxyLists))]
	runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[INF] 当前使用代理 %s ", current))
	runtime.EventsEmit(a.ctx, "status_update", current)

	timeout, err := strconv.Atoi(a.config.Timeout)
	if err != nil {
		a.Debug("Invalid timeout value: %v", err)
	}
	socks, err := net.DialTimeout("tcp", current, time.Duration(timeout)*time.Second)
	if err != nil {
		a.Debug("DialTimeout error: %v", err)
	}

	if err != nil {
		runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[ERR] 连接代理失败 %s ", err.Error()))
		a.handleConnection(conn)
		return
	}
	defer socks.Close()

	var wg sync.WaitGroup
	ioCopy := func(dst io.Writer, src io.Reader) {
		defer wg.Done()
		_, err := io.Copy(dst, src)
		if err != nil {
			runtime.EventsEmit(a.ctx, "log_update", fmt.Sprintf("[ERR] 数据传输失败 %s ", err.Error()))
		}
	}

	wg.Add(2)
	go ioCopy(socks, conn)
	go ioCopy(conn, socks)
	wg.Wait()
}

// 停止任务，释放端口
func (a *App) stopTask() Response {
	if a.config.GetStatus() == 2 {
		// 取消上下文，以停止监听
		cancel()         // 取消监听操作
		listener.Close() // 关闭监听器
		listener = nil   // 清空监听器
		runtime.EventsEmit(a.ctx, "log_update", "[INF] 停止监听服务")
		runtime.EventsEmit(a.ctx, "log_update", "=========================== 任务停止 ============================")
	}
	a.config.SetStatus(0) // 更新任务状态为停止

	// 返回成功响应
	return a.successResponse("监听已成功停止", nil)
}
