package cmd

import (
	"fmt"
	"os"

	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spelens-gud/trunk/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// centerConfigFile 配置文件路径
	centerConfigFile string
	// centerViper Center 服务专用的 viper 实例
	centerViper *viper.Viper
)

var centerCmd = &cobra.Command{
	Use:   "center",
	Short: "启动 Center 服务",
	Long:  `Center 服务负责中心服务器的管理和协调`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Center 配置
		utils.MustNoError(initCenterConfig(), "加载配置文件失败")

		// 从 Center 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(centerViper)

		// 创建日志实例
		log := utils.MustValue(logger.NewLogger(logConfig))
		defer log.Sync()

		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Center 服务运行中...")
	},
}

func init() {
	rootCmd.AddCommand(centerCmd)

	// 定义 Center 服务的配置文件标志
	centerCmd.Flags().StringVar(&centerConfigFile, "config", "", "配置文件路径 (默认为 ./config/center.yaml)")
}

// initCenterConfig 初始化 Center 配置
func initCenterConfig() error {
	// 创建 Center 服务专用的 viper 实例
	centerViper = viper.New()

	if centerConfigFile != "" {
		// 使用命令行指定的配置文件
		centerViper.SetConfigFile(centerConfigFile)
	} else {
		// 查找主目录
		home := utils.MustFuncValue(os.UserHomeDir, "获取用户主目录失败")

		centerViper.AddConfigPath(home)
		centerViper.AddConfigPath(".")
		centerViper.AddConfigPath("./config")
		centerViper.SetConfigType("yaml")
		centerViper.SetConfigName("center")
	}

	// 读取环境变量
	centerViper.AutomaticEnv()

	// 读取配置文件
	if err := centerViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", centerViper.ConfigFileUsed())
	return nil
}
