package cmd

import (
	"fmt"
	"os"

	"github.com/spelens-gud/trunk/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// 全局配置文件路径
	cfgFile string
	// 全局日志级别
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "trunk",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查 version 标志
		versionFlag, err := cmd.Flags().GetBool("version")
		if err != nil {
			return err
		}

		if versionFlag {
			internal.GetVersionInfo()
		}
		return cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志，在这里定义标志并绑定到配置
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为 $HOME/.trunk/trunk.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "日志级别 (debug, info, warn, error)")

	// Cobra 也支持本地标志，只在直接调用此操作时运行
	rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")
}

func initConfig() {
	if cfgFile != "" {
		// 使用命令行指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找主目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("trunk")
	}

	// 读取环境变量
	viper.AutomaticEnv()

	// 如果找到配置文件，则读取它
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "使用配置文件:", viper.ConfigFileUsed())
	}
}
