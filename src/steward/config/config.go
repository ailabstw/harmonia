package config

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gopkg.in/yaml.v3"

	"harmonia.com/steward/config/iconfig"
)

var Config *iconfig.IConfig

func validateGeneralConfig(config *iconfig.IConfig) {
	if config.Notification.Type != "pull" {
		config.Notification.Type = "push"
	}
	zap.L().Debug("", zap.String("notification.type", config.Notification.Type))

	if config.Notification.Type == "push" {
		if config.Notification.StewardServerURI == "" {
			config.Notification.StewardServerURI = "0.0.0.0:9080"
		}
		zap.L().Debug("", zap.String("notification.stewardServerURI", config.Notification.StewardServerURI))
	} else {
		if config.Notification.PullPeriod == 0 {
			config.Notification.PullPeriod = 10
		}
		zap.L().Debug("", zap.Int("notification.pullPeriod", config.Notification.PullPeriod))
	}

	if config.GitUserToken == "" {
		zap.L().Fatal("gitUserToken undefined")
	}
	zap.L().Debug("", zap.String("gitUserToken", config.GitUserToken))

	if config.OperatorGrpcServerURI == "" {
		config.OperatorGrpcServerURI = "localhost:8787"
	}
	zap.L().Debug("", zap.String("operatorGrpcServerURI", config.OperatorGrpcServerURI))
	if config.AppGrpcServerURI == "" {
		config.AppGrpcServerURI = "localhost:7878"
	}
	zap.L().Debug("", zap.String("appGrpcServerURI", config.AppGrpcServerURI))

	checkIsNodeDefined("trainPlanRepo", config.TrainPlanRepo)
}

func initLogConfig(logLevelString string, logPath string) {
	logLevelMap := map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"":      zapcore.InfoLevel, // make the zero value useful
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
		"panic": zapcore.PanicLevel,
		"fatal": zapcore.FatalLevel,
	}
	var logLevel zapcore.Level

	if val, ok := logLevelMap[logLevelString]; ok {
		logLevel = val
	} else {
		panic("invalid log level")
	}

	outputPaths := []string{"stdout"}
	if logPath != "" {
		outputPaths = append(outputPaths, logPath)
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Encoding:         "json",
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "timestamp",
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
	zap.L().Debug("logger init succeeded")
}

func validateAggregatorConfig(config *iconfig.IConfig) {
	if config.AggregatedModelRepo == nil {
		zap.L().Fatal("Need `aggregatorModelRepo` Information.")
	}
	checkIsNodeDefined("aggregatorModelRepo", config.AggregatedModelRepo)

	if config.EdgeModelRepos == nil || len(config.EdgeModelRepos) == 0 {
		zap.L().Fatal("Need `edgeModelRepos` Information.")
	}

	edges := config.EdgeModelRepos
	for index, edge := range edges {
		checkIsNodeDefined(fmt.Sprintf("edgeModelRepos[%d]", index), edge)
	}
}

func validateEdgeConfig(config *iconfig.IConfig) {
	if config.AggregatedModelRepo == nil {
		zap.L().Fatal("Need `aggregatorModelRepo` Information.")
	}
	checkIsNodeDefined("aggregatorModelRepo", config.AggregatedModelRepo)

	zap.L().Debug("edge info...")
	if config.EdgeModelRepo == nil {
		zap.L().Fatal("Need `edgeModelRepo` Information.")
	}
	checkIsNodeDefined("edgeModelRepo", config.EdgeModelRepo)
}

func validateConfig(_config *iconfig.IConfig) {
	var validateConfigByType = map[string]func(*iconfig.IConfig){
		"aggregator": validateAggregatorConfig,
		"edge":       validateEdgeConfig,
	}

	if validateConfigByType[_config.Type] == nil {
		zap.L().Fatal("invalid type")
	}

	validateGeneralConfig(_config)

	zap.L().Debug("", zap.String("operator Type", _config.Type))
	validateConfigByType[_config.Type](_config)
}

func checkIsNodeDefined(fieldName string, node *iconfig.RepoInfo) {
	if node.GitHttpURL == "" {
		zap.L().Fatal(fmt.Sprintf("the information of `%s` node is not completed", fieldName))
	}

	zap.L().Debug(fmt.Sprintf("Repository [%s]: %v", fieldName, node))
}

func init() {
	Config = &iconfig.IConfig{}
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic(fmt.Sprintf("read config file get error: %v", err))
	}

	err = yaml.Unmarshal(yamlFile, &Config)

	if err != nil {
		panic(fmt.Sprintf("unmarshal config yaml get error: %v", err))
	}
	initLogConfig(Config.LogLevel, Config.LogPath)
	validateConfig(Config)
}
