// Command search-service 是从 Ydsz Plane 单体核心独立部署的
// 第二个微服务（S14 Phase-3）。
//
// 职责：
//   - 消费 RabbitMQ 领域事件 → 写入 ES 索引
//   - 提供全文检索能力的 gRPC 接口（ES 主路径 + PG FTS 降级）
//   - 执行全量索引重建（运维触发）
//   - 管理搜索历史 / 收藏
//
// 依赖：
//   - PostgreSQL（read-only，读搜索数据）
//   - Elasticsearch 8.x + IK 分词
//   - RabbitMQ（消费领域事件）
//
// 部署方式：
//   docker run ydsz-plane-search:latest
//   环境变量：DATABASE_URL、ES_URLS、RABBITMQ_URL、GRPC_PORT
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

	searchv1 "github.com/njydsz/ydsz-plane/api/proto/search/v1"
	"github.com/njydsz/ydsz-plane/internal/application/search"
)

const (
	defaultGRPCPort   = "9091" // 与 gRPC notification-service 区分
	defaultHealthPort = "8081"
	shutdownTimeout   = 15 * time.Second
)

func main() {
	grpcPort := env("GRPC_PORT", defaultGRPCPort)
	healthPort := env("HEALTH_PORT", defaultHealthPort)

	log.SetPrefix("[search-service] ")
	log.Printf("starting (grpc=%s health=%s)", grpcPort, healthPort)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 1. 初始化搜索服务
	// TODO: 接入真实 DB/ES/RabbitMQ 连接池
	var searchSvc *search.Service   // = search.NewService(dbPool)
	var searchIndexer *search.Indexer // = search.NewIndexer(dbPool)

	// 2. gRPC Server
	grpcServer := grpc.NewServer()
	var grpcSvc *search.GRPCService
	if searchSvc != nil && searchIndexer != nil {
		grpcSvc = search.NewGRPCService(searchSvc, searchIndexer)
		searchv1.RegisterSearchServiceServer(grpcServer, grpcSvc)
		if grpcSvc != nil {
			_ = grpcSvc
		}
	}

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("ydszplane.search.v1.SearchService",
		grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", grpcPort))
	if err != nil {
		log.Fatalf("gRPC listen: %v", err)
	}

	// 3. HTTP Health + Metrics
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ready":true}`))
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("# metrics endpoint (TODO)"))
	})

	httpServer := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%s", healthPort), Handler: mux}
	go func() {
		log.Printf("health server listening on :%s", healthPort)
		_ = httpServer.ListenAndServe()
	}()

	// 4. 优雅关闭
	go func() {
		<-ctx.Done()
		log.Println("shutdown signal received, draining...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()
		grpcServer.GracefulStop()
		_ = httpServer.Shutdown(shutdownCtx)
		log.Println("search-service stopped")
	}()

	log.Println("search-service ready")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC serve: %v", err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
