package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"ride-hailing/admin-server/core/internal"
	"ride-hailing/admin-server/global"
	"ride-hailing/admin-server/pkg/config"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func getConfigPath() string {
	if p := os.Getenv(internal.ConfigEnv); p != "" {
		return p
	}
	return internal.ConfigDefaultFile
}

func Viper() *viper.Viper {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	nacosConfigDir := filepath.Join(configDir, "configs")
	if _, err := os.Stat(filepath.Join(nacosConfigDir, "nacos.yaml")); err == nil {
		data, source, err := config.LoadConfig(nacosConfigDir)
		if err == nil {
			fmt.Printf("配置来源: %s\n", source)
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
				panic(fmt.Errorf("fatal error parse config: %w", err))
			}
			if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
				panic(fmt.Errorf("fatal error unmarshal config: %w", err))
			}
			global.GVA_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
			return v
		}
		fmt.Printf("Nacos config load failed, fallback local: %v\n", err)
	}

	config := configPath
	v := viper.New()
	v.SetConfigFile(config)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
			fmt.Println(err)
		}
	})
	if err = v.Unmarshal(&global.GVA_CONFIG); err != nil {
		panic(fmt.Errorf("fatal error unmarshal config: %w", err))
	}

	global.GVA_CONFIG.AutoCode.Root, _ = filepath.Abs("..")
	return v
}
