package cmd

import (
	"fmt"
	"os"

	"github.com/spelens-gud/trunk/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 配置文件路径
	centerConfigFile string
)

var centerCmd = &cobra.Command{
	Use:   "center",
	Short: "启动 Center 服务",
	Long:  `Center 服务负责中心服务器的管理和协调`,
	Run: func(cmd *cobra.Command, args []string) {
		// 从 viper 加载日志配置
		logConfig := logger.LoadConfigFromViper()

		// 创建日志实例
		log, err := logger.NewLogger(logConfig)
		if err != nil {
			panic(err)
		}
		defer log.Sync()

		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("服务运行中...")
	},
}

func init() {
	rootCmd.AddCommand(centerCmd)
	cobra.OnInitialize(initCenterConfig)

	// 全局标志，在这里定义标志并绑定到配置
	rootCmd.PersistentFlags().StringVar(&centerConfigFile, "center_config", "", "配置文件路径 (默认为 $HOME/.trunk/center.yaml)")
}

// initCenterConfig 初始化配置
func initCenterConfig() {
	if centerConfigFile != "" {
		// 使用命令行指定的配置文件
		viper.SetConfigFile(centerConfigFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("center")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}
