package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// friendConfigFile 配置文件路径
	friendConfigFile string
	// friendViper Friend 服务专用的 viper 实例
	friendViper *viper.Viper
)

var friendCmd = &cobra.Command{
	Use:   "friend",
	Short: "启动 Friend 服务",
	Long:  `Friend 服务负责好友系统的管理`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Friend 配置
		utils.MustNoError(initFriendConfig(), "加载配置文件失败")

		// 从 Friend 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(friendViper)

		// 创建日志实例
		log := utils.MustValue(logger.NewLogger(logConfig))
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

		if err := gracefulShutdownFriendServer(shutdownCtx); err != nil {
			log.Infof("客户端关闭失败: %v", err)
		} else {
			log.Infof("客户端已成功关闭")
		}

	},
}

func init() {
	rootCmd.AddCommand(friendCmd)

	// 定义 Friend 服务的配置文件标志
	friendCmd.Flags().StringVar(&friendConfigFile, "config", "", "配置文件路径 (默认为 ./config/friend.yaml)")
}

// initFriendConfig 初始化 Friend 配置
func initFriendConfig() error {
	// 创建 Friend 服务专用的 viper 实例
	friendViper = viper.New()

	if friendConfigFile != "" {
		// 使用命令行指定的配置文件
		friendViper.SetConfigFile(friendConfigFile)
	} else {
		// 查找主目录
		home := utils.MustFuncValue(os.UserHomeDir, "获取用户主目录失败")

		friendViper.AddConfigPath(home)
		friendViper.AddConfigPath(".")
		friendViper.AddConfigPath("./config")
		friendViper.SetConfigType("yaml")
		friendViper.SetConfigName("friend")
	}

	// 读取环境变量
	friendViper.AutomaticEnv()

	// 读取配置文件
	if err := friendViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", friendViper.ConfigFileUsed())
	return nil
}

// gracefulShutdownCenterServer 优雅关闭中心服
func gracefulShutdownFriendServer(ctx context.Context) error {
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
