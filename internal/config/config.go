package config

import (
	"github.com/spf13/viper"
	"github.com/xylonx/zapx"
	"go.uber.org/zap"
)

var Config *Setting

func Setup(cfgFile string) error {
	Config = new(Setting)

	v := viper.New()

	v.SetConfigFile(cfgFile)

	if err := v.ReadInConfig(); err != nil {
		zapx.Error("read config failed", zap.Error(err))
		return err
	}

	if err := v.Unmarshal(Config); err != nil {
		zapx.Error("unmarshal config file failed", zap.Error(err))
		return err
	}

	return nil
}
