package grpc

import (
	"testing"
	"time"

	"github.com/spelens-gud/logger"
)

func TestGrpcNetClient_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &GrpcNetClient{
		cnf: &ClientConfig{
			Name:             "test-client",
			Host:             "localhost:50051",
			KeepAliveTime:    10 * time.Second,
			KeepAliveTimeout: 3 * time.Second,
		},
		log: log,
	}

	client.New()

	if client.stopChan == nil {
		t.Error("stopChan 未初始化")
	}

	if !client.isStop {
		t.Error("isStop 应该为 true")
	}
}

func TestGrpcNetClient_IsConnected(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &GrpcNetClient{
		cnf: &ClientConfig{
			Name: "test-client",
			Host: "localhost:50051",
		},
		log:    log,
		isStop: true,
	}

	client.New()

	if client.IsConnected() {
		t.Error("期望未连接，但返回已连接")
	}
}

func TestGrpcNetClient_GetReconnectCount(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &GrpcNetClient{
		cnf: &ClientConfig{
			Name: "test-client",
			Host: "localhost:50051",
		},
		log:            log,
		reconnectCount: 3,
	}

	count := client.GetReconnectCount()
	if count != 3 {
		t.Errorf("期望重连次数 = 3, 实际 = %d", count)
	}
}
