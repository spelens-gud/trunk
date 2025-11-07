package cmd

import (
	"os"

	"github.com/spelens-gud/trunk/internal/version"
	"github.com/spf13/cobra"
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
			version.PrintVersion()

			return nil
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
	// Cobra 也支持本地标志，只在直接调用此操作时运行
	rootCmd.Flags().BoolP("version", "v", false, "显示版本信息")
}
