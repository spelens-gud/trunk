package quic

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/spelens-gud/logger"
)

func generateTestTLSConfig() *tls.Config {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, _ := tls.X509KeyPair(certPEM, keyPEM)
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-trunk"},
	}
}

func TestServerConfig_GetMethods(t *testing.T) {
	config := &ServerConfig{
		MaxConnections: 100,
		IdleTimeout:    30 * time.Second,
	}

	if config.GetMaxConnections() != 100 {
		t.Errorf("期望 MaxConnections = 100, 实际 = %d", config.GetMaxConnections())
	}

	if config.GetIdleTimeout() != 30*time.Second {
		t.Errorf("期望 IdleTimeout = 30s, 实际 = %v", config.GetIdleTimeout())
	}
}

func TestQuicNetServer_New(t *testing.T) {
	log, _ := logger.NewLogger(&logger.Config{
		Level:   "info",
		Console: true,
	})

	server := &NetQuicServer{
		cnf: &ServerConfig{
			Name:      "test-server",
			Ip:        "127.0.0.1",
			Port:      8443,
			TLSConfig: generateTestTLSConfig(),
		},
		log: log,
	}

	server.New()

	if server.stopChan == nil {
		t.Error("stopChan 未初始化")
	}
}
