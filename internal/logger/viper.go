package logger

import "github.com/spf13/viper"

// LoadConfigFromViper 从 viper 加载日志配置（从根级别读取）
func LoadConfigFromViper() *Config {
	return &Config{
		ServiceName:      viper.GetString("service_name"),
		Environment:      viper.GetString("environment"),
		Level:            viper.GetString("log_level"),
		Console:          viper.GetBool("console"),
		File:             viper.GetBool("file"),
		FilePath:         viper.GetString("file_path"),
		FileName:         viper.GetString("file_name"),
		MaxSize:          viper.GetInt("max_size"),
		MaxAge:           viper.GetInt("max_age"),
		MaxBackups:       viper.GetInt("max_backups"),
		Compress:         viper.GetBool("compress"),
		EnableCaller:     viper.GetBool("enable_caller"),
		EnableStacktrace: viper.GetBool("enable_stacktrace"),
	}
}
