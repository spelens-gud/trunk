package grpc

import (
	"testing"
	"time"

	"github.com/spelens-gud/logger"
)

func TestServerConfig_GetMethods(t *testing.T) {
	config := &ServerConfig{
		MaxConnections:       100,
		MaxConcurrentStreams: 50,
	}

	if config.GetMaxConnections() != 100 {
		t.Errorf("期望 MaxConnections = 100, 实际 = %d", config.GetMaxConnections())
	}

	if config.GetMaxConcurrentStreams() != 50 {
		t.Errorf("期望 MaxConcurrentStreams = 50, 实际 = %d", config.GetMaxConcurrentStreams())
	}
}

func TestGrpcNetServer_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &GrpcNetServer{
		cnf: &ServerConfig{
			Name:                 "test-server",
			Ip:                   "127.0.0.1",
			Port:                 50051,
			MaxConnections:       100,
			MaxConcurrentStreams: 50,
			KeepAliveTime:        10 * time.Second,
			KeepAliveTimeout:     3 * time.Second,
		},
		log: log,
	}

	server.New()

	if server.stopChan == nil {
		t.Error("stopChan 未初始化")
	}

	if server.server == nil {
		t.Error("server 未初始化")
	}
}

func TestGrpcNetServer_Stats(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &GrpcNetServer{
		cnf: &ServerConfig{
			Name: "test-server",
			Ip:   "127.0.0.1",
			Port: 50052,
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
