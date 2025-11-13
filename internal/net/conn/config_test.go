package conn

import (
	"errors"
	"testing"
	"time"
)

// TestNetConfig_Validate 测试配置验证
func TestNetConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NetConfig[*mockConn]
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置",
			config: NetConfig[*mockConn]{
				OnWrite: func(conn *mockConn, raw []byte) error { return nil },
				OnRead:  func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
				OnData:  func(conn IConn, raw []byte) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "缺少OnWrite",
			config: NetConfig[*mockConn]{
				OnRead: func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
				OnData: func(conn IConn, raw []byte) error { return nil },
			},
			wantErr: true,
			errMsg:  "OnWrite回调函数未设置",
		},
		{
			name: "缺少OnRead",
			config: NetConfig[*mockConn]{
				OnWrite: func(conn *mockConn, raw []byte) error { return nil },
				OnData:  func(conn IConn, raw []byte) error { return nil },
			},
			wantErr: true,
			errMsg:  "OnRead回调函数没有设置",
		},
		{
			name: "缺少OnData",
			config: NetConfig[*mockConn]{
				OnWrite: func(conn *mockConn, raw []byte) error { return nil },
				OnRead:  func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
			},
			wantErr: true,
			errMsg:  "OnData回调函数没有设置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestNetConfig_GetWriteTimeout 测试获取写超时
func TestNetConfig_GetWriteTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{
			name:    "使用默认超时",
			timeout: 0,
			want:    DefaultWriteTimeOut,
		},
		{
			name:    "使用自定义超时",
			timeout: time.Second * 10,
			want:    time.Second * 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &NetConfig[*mockConn]{
				WriteTimeout: tt.timeout,
			}
			got := cfg.GetWriteTimeout()
			if got != tt.want {
				t.Errorf("GetWriteTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNetConfig_GetReadTimeout 测试获取读超时
func TestNetConfig_GetReadTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{
			name:    "使用默认超时",
			timeout: 0,
			want:    DefaultReadTimeOut,
		},
		{
			name:    "使用自定义超时",
			timeout: time.Minute * 10,
			want:    time.Minute * 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &NetConfig[*mockConn]{
				ReadTimeout: tt.timeout,
			}
			got := cfg.GetReadTimeout()
			if got != tt.want {
				t.Errorf("GetReadTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNetConfig_OnClose 测试OnClose回调
func TestNetConfig_OnClose(t *testing.T) {
	mc := newMockConn()
	called := false

	cfg := NetConfig[*mockConn]{
		OnWrite: func(conn *mockConn, raw []byte) error { return nil },
		OnRead:  func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
		OnData:  func(conn IConn, raw []byte) error { return nil },
		OnClose: func(conn *mockConn) error {
			called = true
			return nil
		},
	}

	// 验证OnClose可以被调用
	err := cfg.OnClose(mc)
	if err != nil {
		t.Errorf("OnClose() error = %v", err)
	}

	if !called {
		t.Error("OnClose回调应该被调用")
	}
}

// TestNetConfig_OnCloseError 测试OnClose返回错误
func TestNetConfig_OnCloseError(t *testing.T) {
	mc := newMockConn()
	expectedErr := errors.New("关闭错误")

	cfg := NetConfig[*mockConn]{
		OnWrite: func(conn *mockConn, raw []byte) error { return nil },
		OnRead:  func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
		OnData:  func(conn IConn, raw []byte) error { return nil },
		OnClose: func(conn *mockConn) error {
			return expectedErr
		},
	}

	err := cfg.OnClose(mc)
	if err != expectedErr {
		t.Errorf("OnClose() error = %v, want %v", err, expectedErr)
	}
}

// BenchmarkNetConfig_Validate 基准测试：配置验证
func BenchmarkNetConfig_Validate(b *testing.B) {
	cfg := NetConfig[*mockConn]{
		OnWrite: func(conn *mockConn, raw []byte) error { return nil },
		OnRead:  func(conn *mockConn) (int, []byte, error) { return 0, nil, nil },
		OnData:  func(conn IConn, raw []byte) error { return nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

// BenchmarkNetConfig_GetWriteTimeout 基准测试：获取写超时
func BenchmarkNetConfig_GetWriteTimeout(b *testing.B) {
	cfg := NetConfig[*mockConn]{
		WriteTimeout: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.GetWriteTimeout()
	}
}

// BenchmarkNetConfig_GetReadTimeout 基准测试：获取读超时
func BenchmarkNetConfig_GetReadTimeout(b *testing.B) {
	cfg := NetConfig[*mockConn]{
		ReadTimeout: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.GetReadTimeout()
	}
}
