package cmd

import (
	"fmt"
	"os"

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
		if err := initGateConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "加载配置文件失败: %v\n", err)
			os.Exit(1)
		}

		// 从 Gate 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(gateViper)

		// 创建日志实例
		log, err := logger.NewLogger(logConfig)
		if err != nil {
			panic(err)
		}
		defer log.Sync()

		// 记录服务启动
		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Gate 服务运行中...")
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

	if gateConfigFile != "" {
		// 使用命令行指定的配置文件
		gateViper.SetConfigFile(gateConfigFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}

		gateViper.AddConfigPath(home)
		gateViper.AddConfigPath(".")
		gateViper.AddConfigPath("./config")
		gateViper.SetConfigType("yaml")
		gateViper.SetConfigName("gate")
	}

	// 读取环境变量
	gateViper.AutomaticEnv()

	// 读取配置文件
	if err := gateViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", gateViper.ConfigFileUsed())
	return nil
}
