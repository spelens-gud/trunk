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
	// nodeConfigFile 配置文件路径
	nodeConfigFile string
	// nodeViper Node 服务专用的 viper 实例
	nodeViper *viper.Viper
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "启动 Node 服务",
	Long:  `Node 服务负责节点管理和分布式协调`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Node 配置
		assert.MustCall0E(initNodeConfig, "加载配置文件失败")

		// 从 Node 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(nodeViper)

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

		assert.ShouldCall1E(gracefulShutdownNodeServer, shutdownCtx, "客户端关闭失败")
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// 定义 Node 服务的配置文件标志
	nodeCmd.Flags().StringVar(&nodeConfigFile, "config", "", "配置文件路径 (默认为 ./config/node.yaml)")
}

// initNodeConfig 初始化 Node 配置
func initNodeConfig() error {
	// 创建 Node 服务专用的 viper 实例
	nodeViper = viper.New()

	assert.Then(nodeConfigFile != "").Do(func() {
		// 使用命令行指定的配置文件
		nodeViper.SetConfigFile(nodeConfigFile)
	}).Else(func() {
		// 查找主目录
		home := assert.MustCall0RE(os.UserHomeDir, "获取用户主目录失败")
		nodeViper.AddConfigPath(home)
		nodeViper.AddConfigPath(".")
		nodeViper.AddConfigPath("./config")
		nodeViper.SetConfigType("yaml")
		nodeViper.SetConfigName("node")
	})

	// 读取环境变量
	nodeViper.AutomaticEnv()

	// 读取配置文件
	assert.MustCall0E(nodeViper.ReadInConfig, "读取配置文件失败")

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", nodeViper.ConfigFileUsed())
	return nil
}

// gracefulShutdownNodeServer 优雅关闭中心服
func gracefulShutdownNodeServer(ctx context.Context) error {
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
