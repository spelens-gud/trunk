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

		// 记录服务启动
		log.Infof("服务的配置文件：%v", logConfig)

		// 这里添加你的服务逻辑
		log.Info("Fight 服务运行中...")
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
