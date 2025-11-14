package tars

import (
	"testing"

	"github.com/spelens-gud/logger"
)

func TestServerConfig_GetMethods(t *testing.T) {
	config := &ServerConfig{
		MaxConnections: 100,
	}

	if config.GetMaxConnections() != 100 {
		t.Errorf("期望 MaxConnections = 100, 实际 = %d", config.GetMaxConnections())
	}
}

func TestTarsNetServer_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &TarsNetServer{
		cnf: &ServerConfig{
			Name:     "test-server",
			Ip:       "127.0.0.1",
			Port:     10000,
			Protocol: "tcp",
		},
		log: log,
	}

	server.New()

	if server.stopChan == nil {
		t.Error("stopChan 未初始化")
	}

	if server.comm == nil {
		t.Error("comm 未初始化")
	}
}

func TestTarsNetServer_Stats(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &TarsNetServer{
		cnf: &ServerConfig{
			Name: "test-server",
			Ip:   "127.0.0.1",
			Port: 10001,
		},
		log:           log,
		connCount:     5,
		totalAccepted: 100,
		totalRejected: 10,
	}

	server.New()

	stats := server.GetStats()
	if stats.CurrentConnections != 5 {
		t.Errorf("期望 CurrentConnections = 5, 实际 = %d", stats.CurrentConnections)
	}
}
