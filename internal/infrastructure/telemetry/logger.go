// Package telemetry 初始化结构化日志（zap），后续将扩展指标与链路追踪。
// 日志字段遵循约定：ts/level/msg/request_id/tenant/user/trace_id。
package telemetry

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 根据配置值构建 zap logger。
// format 取值："json"（生产环境）或 "console"（开发环境）。
// 生产环境默认不开启颜色编码，便于日志采集系统解析。
func NewLogger(level, format string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.LevelKey = "level"
	encCfg.MessageKey = "msg"

	var encoder zapcore.Encoder
	if format == "console" {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), lvl)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
