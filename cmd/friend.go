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

		// 记录服务启动
		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Friend 服务运行中...")
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
