package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spelens-gud/trunk/internal/assert"
	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// gateConfigFile 配置文件路径
	gateConfigFile string
	// gateViper Gate 服务专用的 viper 实例
	gateViper *viper.Viper
)

var gateCmd = &cobra.Command{
	Use:   "gate",
	Short: "启动 Gate 服务",
	Long:  `Gate 服务负责网关和连接管理`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Gate 配置
		assert.MustCall0E(initGateConfig, "加载配置文件失败")

		// 从 Gate 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(gateViper)

		// 创建日志实例
		log := assert.MustCall1RE(logger.NewLogger, logConfig, "创建日志实例失败")
		assert.SetLogger(log) // 注入assert模块
		defer log.Sync()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 启动服务
		go func() {
			// TODO
			// 1. 初始化服务
			// 2. 启动服务
		}()

		select {
		case sig := <-sigChan:
			log.Infof("收到信号 %v，正在关闭客户端...", sig)
			cancel()
		case <-ctx.Done():
			log.Infof("客户端上下文已取消")
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		assert.ShouldCall1E(gracefulShutdownGateServer, shutdownCtx, "客户端关闭失败")
	},
}

func init() {
	rootCmd.AddCommand(gateCmd)

	// 定义 Gate 服务的配置文件标志
	gateCmd.Flags().StringVar(&gateConfigFile, "config", "", "配置文件路径 (默认为 ./config/gate.yaml)")
}

// initGateConfig 初始化 Gate 配置
func initGateConfig() error {
	// 创建 Gate 服务专用的 viper 实例
	gateViper = viper.New()

	assert.Then(gateConfigFile != "").Do(func() {
		// 使用命令行指定的配置文件
		gateViper.SetConfigFile(gateConfigFile)
	}).Else(func() {
		// 查找主目录
		home := assert.MustCall0RE(os.UserHomeDir, "获取用户主目录失败")
		gateViper.AddConfigPath(home)
		gateViper.AddConfigPath(".")
		gateViper.AddConfigPath("./config")
		gateViper.SetConfigType("yaml")
		gateViper.SetConfigName("gate")
	})

	// 读取环境变量
	gateViper.AutomaticEnv()

	// 读取配置文件
	assert.MustCall0E(gateViper.ReadInConfig, "读取配置文件失败")

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", gateViper.ConfigFileUsed())
	return nil
}

// gracefulShutdownGateServer 优雅关闭中心服
func gracefulShutdownGateServer(ctx context.Context) error {
	log.Println("开始优雅关闭客户端...")

	// TODO: 实现客户端关闭资源释放

	select {
	case <-time.After(1 * time.Second):
		log.Println("客户端关闭完成")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
