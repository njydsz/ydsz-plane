// Package dashboard — 仪表盘实时广播封装。
//
// 提供风险告警 WebSocket 广播能力，worker 定时检测后推送至全工作空间。
package dashboard

import (
	"context"
	"encoding/json"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
)

// BroadcastRiskAlert 将风险告警广播到指定工作空间的全部 WebSocket 客户端。
//
// 消息 payload 格式：{"type":"risk_alert","data":<alert JSON>}
// 由 risk detection worker 在检测到新告警后调用。
func BroadcastRiskAlert(ctx context.Context, hub *ws.Hub, wsID int64, alert RiskAlert) error {
	data, err := json.Marshal(alert)
	if err != nil {
		return err
	}
	msg := ws.Message{
		Type: "risk_alert",
		Data: json.RawMessage(data),
	}
	return hub.Publish(ctx, wsID, msg)
}
