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
	nodeConfigFile string
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "启动 Node 服务",
	Long:  `Node 服务负责节点管理和分布式协调`,
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
		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Node 服务运行中...")
	},
}

func init() {
	rootCmd.AddCommand(fightCmd)
	cobra.OnInitialize(initNodeConfig)

	// 全局标志，在这里定义标志并绑定到配置
	rootCmd.PersistentFlags().StringVar(&fightConfigFile, "node_config", "", "配置文件路径 (默认为 $HOME/.trunk/center.yaml)")
}

// initNodeConfig 初始化配置
func initNodeConfig() {
	if nodeConfigFile != "" {
		// 使用命令行指定的配置文件
		viper.SetConfigFile(nodeConfigFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("node")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}
