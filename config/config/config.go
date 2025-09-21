package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// viper读取配置文件
func InitViper() {
	// 设置配置文件名称
	viper.SetConfigName("config")
	// 设置配置文件类型
	viper.SetConfigType("yaml")
	// 添加配置文件搜索路径
	viper.AddConfigPath(".")         // 当前目录
	viper.AddConfigPath("./config/") // 其他搜索路径

	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("log.level", "info")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 自动绑定环境变量（将以APP_开头的环境变量绑定到配置键）
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
}
