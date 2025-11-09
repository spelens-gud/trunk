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
	// fightConfigFile 配置文件路径
	fightConfigFile string
	// fightViper Fight 服务专用的 viper 实例
	fightViper *viper.Viper
)

var fightCmd = &cobra.Command{
	Use:   "fight",
	Short: "启动 Fight 服务",
	Long:  `Fight 服务负责战斗逻辑的处理`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Fight 配置
		utils.MustNoError(initFightConfig(), "加载配置文件失败")

		// 从 Fight 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(fightViper)

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

		if err := gracefulShutdownFightServer(shutdownCtx); err != nil {
			log.Infof("客户端关闭失败: %v", err)
		} else {
			log.Infof("客户端已成功关闭")
		}
	},
}

func init() {
	rootCmd.AddCommand(fightCmd)

	// 定义 Fight 服务的配置文件标志
	fightCmd.Flags().StringVar(&fightConfigFile, "config", "", "配置文件路径 (默认为 ./config/fight.yaml)")
}

// initFightConfig 初始化 Fight 配置
func initFightConfig() error {
	// 创建 Fight 服务专用的 viper 实例
	fightViper = viper.New()

	if fightConfigFile != "" {
		// 使用命令行指定的配置文件
		fightViper.SetConfigFile(fightConfigFile)
	} else {
		// 查找主目录
		home := utils.MustFuncValue(os.UserHomeDir, "获取用户主目录失败")

		fightViper.AddConfigPath(home)
		fightViper.AddConfigPath(".")
		fightViper.AddConfigPath("./config")
		fightViper.SetConfigType("yaml")
		fightViper.SetConfigName("fight")
	}

	// 读取环境变量
	fightViper.AutomaticEnv()

	// 读取配置文件
	if err := fightViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", fightViper.ConfigFileUsed())
	return nil
}

// gracefulShutdownCenterServer 优雅关闭中心服
func gracefulShutdownFightServer(ctx context.Context) error {
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
