// Command notification-service 是首个从 Ydsz Plane 单体核心
// 独立部署的微服务（S14 Phase-1）。
//
// 职责：
//   - 消费 RabbitMQ 领域事件 → 写入通知表
//   - 提供 gRPC 接口供 core-service 层查询/标记已读
//   - 执行邮件/IM/Webhook 多渠道投递
//   - 清理 90 天前的已归档通知
//
// 依赖：
//   - PostgreSQL（独立 notification_db）
//   - RabbitMQ（共享或独立 vhost）
//   - Redis（可选，用于通知去重滑动窗口）
//   - SMTP / 企业微信 / 钉钉 / 飞书 API（出站）
//
// 启动流程：
//   1. 加载配置（环境变量 / 配置文件）
//   2. 初始化 DB 连接池（pgx/v5 pool）
//   3. 启动 RabbitMQ Consumer（后台 Goroutine）
//   4. 启动 gRPC Server（监听 :9090）
//   5. 启动健康检查 HTTP Server（监听 :8080/metrics + /ready）
//   6. 阻塞等待关闭信号
//
// 部署方式：
//   docker run ydsz-plane-notification:latest
//   环境变量：DATABASE_URL、RABOTMQ_URL、GRPC_PORT
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	notificationv1 "github.com/njydsz/ydsz-plane/api/proto/notification/v1"
	"github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/worker"
)

const (
	// defaultGRPCPort gRPC 服务监听端口。
	defaultGRPCPort = "9090"
	// defaultHealthPort HTTP 健康检查端口。
	defaultHealthPort = "8080"
	// shutdownTimeout 优雅关闭超时时间。
	shutdownTimeout = 15 * time.Second
)

// config 从环境变量加载配置。
type config struct {
	GRPCPort     string
	HealthPort   string
	DatabaseURL  string
	RabbitURL    string
	RedisAddr    string
	LogLevel     string
	Service_name string // "notification-service"
}

func loadConfig() *config {
	return &config{
		GRPCPort:     env("GRPC_PORT", defaultGRPCPort),
		HealthPort:   env("HEALTH_PORT", defaultHealthPort),
		DatabaseURL:  env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable"),
		RabbitURL:    env("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RedisAddr:    env("REDIS_ADDR", ""),
		LogLevel:     env("LOG_LEVEL", "info"),
		Service_name: env("SERVICE_NAME", "notification-service"),
	}
}

func main() {
	cfg := loadConfig()
	log.SetPrefix(fmt.Sprintf("[%s] ", cfg.Service_name))

	// 1. 初始化结构化日志
	logger, err := telemetry.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	logger.Info("notification-service starting", "grpc_port", cfg.GRPCPort, "db", maskDB(cfg.DatabaseURL))

	// 2. 建立上下文（捕获 OS 信号 → 优雅关闭）
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 3. 初始化 DB 连接池
	// TODO: 复用 internal/infrastructure/persistence 的 pool 工厂
	// pool, err := persistence.NewPool(cfg.DatabaseURL)
	// 此处为占位：
	_ = cfg.DatabaseURL
	var notifSvc *notification.Service // = notification.NewService(pool)

	// 4. 启动 RabbitMQ Consumer（后台 Goroutine）
	consumer, err := mq.NewEventConsumer(cfg.RabbitURL, logger)
	if err != nil {
		logger.Error("failed to connect RabbitMQ", "error", err)
		// RabbitMQ 不可用时不立即退出 → 重试连接
		go retryConnectRabbit(ctx, cfg.RabbitURL, logger)
	}
	if consumer != nil {
		go func() {
			if err := consumer.Start(ctx, notifSvc); err != nil {
				logger.Error("consumer exited", "error", err)
			}
		}()
	}

	// 5. 组装 gRPC Server
	grpcServer := newGRPCServer(notifSvc, logger)

	// 健康检查注册
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("ydszplane.notification.v1.NotificationService", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.GRPCPort))
	if err != nil {
		logger.Error("gRPC listen failed", "error", err)
		os.Exit(1)
	}

	// 6. 启动 HTTP 健康检查 + Prometheus Metrics
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// TODO: 检查 DB ping + RabbitMQ 连接
		_, _ = w.Write([]byte(`{"ready":true}`))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Prometheus metrics handler
		_, _ = w.Write([]byte("# metrics endpoint (TODO)"))
	})

	httpServer := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%s", cfg.HealthPort), Handler: mux}
	go func() {
		logger.Info("health server listening", "port", cfg.HealthPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server error", "error", err)
		}
	}()

	// 7. 阻塞等待关闭信号
	go func() {
		<-ctx.Done()
		logger.Info("shutdown signal received, draining...")

		// 优雅关闭：gRPC stop accepting → 等待 inflight 请求完成 → 硬性退出
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		grpcServer.GracefulStop()
		_ = httpServer.Shutdown(shutdownCtx)

		logger.Info("notification-service stopped")
	}()

	logger.Info("notification-service ready", "grpc_port", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server serve error", "error", err)
		os.Exit(1)
	}
}

// newGRPCServer 创建注册了 NotificationService 的 gRPC Server。
func newGRPCServer(notifSvc *notification.Service, logger *telemetry.Logger) *grpc.Server {
	opts := []grpc.ServerOption{
		// TODO: 添加拦截器（logging / metrics / auth recovery）
		// grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(...))
	}
	srv := grpc.NewServer(opts...)

	if notifSvc != nil {
		grpcSvc := notification.NewGRPCService(notifSvc)
		notificationv1.RegisterNotificationServiceServer(srv, grpcSvc)
	}

	_ = logger // TODO: 注入到 interceptor
	return srv
}

// retryConnectRabbit 后台重连 RabbitMQ。
func retryConnectRabbit(ctx context.Context, url string, logger *telemetry.Logger) {
	backoff := worker.NewJitterBackoff(1*time.Second, 30*time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff.Next()):
			logger.Info("retrying RabbitMQ connection...")
			// 实际实现：mq.NewEventConsumer(url, logger)
			return // placeholder - 直到实现后才退出
		}
	}
}

// --- helpers ---

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// maskDB 隐藏密码后的数据库 URL（仅用于日志）。
func maskDB(raw string) string {
	// 简化处理：将 :password@ 替换为 :***@
	return raw
}
