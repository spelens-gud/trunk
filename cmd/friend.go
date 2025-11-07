package cmd

import (
	"fmt"
	"os"

	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var friendCmd = &cobra.Command{
	Use:   "friend",
	Short: "启动 Friend 服务",
	Long:  `Friend 服务负责好友系统的管理`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 如果没有指定配置文件，使用服务默认配置
		if cfgFile == "" {
			viper.SetConfigFile("./config/friend.yaml")
			if err := viper.ReadInConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "读取配置文件失败: %v\n", err)
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 从 viper 加载日志配置
		logConfig := logger.LoadConfigFromViper()

		// 创建日志实例
		log, err := logger.NewLogger(logConfig)
		if err != nil {
			panic(err)
		}
		defer log.Sync()

		// 记录服务启动
		log.Info("Friend 服务启动",
			zap.String("service", logConfig.ServiceName),
			zap.String("environment", logConfig.Environment),
		)

		// 这里添加你的服务逻辑
		log.Info("Friend 服务运行中...")
	},
}

func init() {
	rootCmd.AddCommand(friendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// friendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// friendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
