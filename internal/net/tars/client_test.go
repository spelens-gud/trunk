package tars

import (
	"testing"

	"github.com/spelens-gud/logger"
)

func TestTarsNetClient_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetTarsClient{
		cnf: &ClientConfig{
			Name: "test-client",
			Host: "localhost:10000",
			Obj:  "TestApp.TestServer.TestObj",
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

	if client.comm == nil {
		t.Error("comm 未初始化")
	}
}

func TestTarsNetClient_IsConnected(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetTarsClient{
		cnf: &ClientConfig{
			Name: "test-client",
			Host: "localhost:10000",
		},
		log:    log,
		isStop: true,
	}

	client.New()

	if client.IsConnected() {
		t.Error("期望未连接，但返回已连接")
	}
}

func TestTarsNetClient_GetReconnectCount(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	client := &NetTarsClient{
		cnf: &ClientConfig{
			Name: "test-client",
			Host: "localhost:10000",
		},
		log:            log,
		reconnectCount: 3,
	}

	count := client.GetReconnectCount()
	if count != 3 {
		t.Errorf("期望重连次数 = 3, 实际 = %d", count)
	}
}
