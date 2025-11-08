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
		utils.MustNoError(initNodeConfig(), "加载配置文件失败")

		// 从 Node 专用的 viper 实例加载日志配置
		logConfig := logger.LoadConfigFromViper(nodeViper)

		// 创建日志实例
		log := utils.MustValue(logger.NewLogger(logConfig))
		defer log.Sync()

		// 记录服务启动
		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Node 服务运行中...")
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

	if nodeConfigFile != "" {
		// 使用命令行指定的配置文件
		nodeViper.SetConfigFile(nodeConfigFile)
	} else {
		// 查找主目录
		home := utils.MustFuncValue(os.UserHomeDir, "获取用户主目录失败")

		nodeViper.AddConfigPath(home)
		nodeViper.AddConfigPath(".")
		nodeViper.AddConfigPath("./config")
		nodeViper.SetConfigType("yaml")
		nodeViper.SetConfigName("node")
	}

	// 读取环境变量
	nodeViper.AutomaticEnv()

	// 读取配置文件
	if err := nodeViper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "使用配置文件: %s\n", nodeViper.ConfigFileUsed())
	return nil
}
