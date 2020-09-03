package iconfig

type IConfig struct {
	Type                  string `yaml:"type"`
	OperatorGrpcServerURI string `yaml:"operatorGrpcServerURI,omitempty"`
	AppGrpcServerURI      string `yaml:"appGrpcServerURI,omitempty"`
	LogLevel              string `yaml:"logLevel"`
	LogPath               string `yaml:"logPath,omitempty"`

	Notification NotificationInfo `yaml:"notification"`

	// Git Associate
	GitUserToken        string      `yaml:"gitUserToken"`
	AggregatedModelRepo *RepoInfo   `yaml:"aggregatorModelRepo,omitempty"`
	EdgeModelRepos      []*RepoInfo `yaml:"edgeModelRepos,omitempty"`
	EdgeModelRepo       *RepoInfo   `yaml:"edgeModelRepo,omitempty"`
	TrainPlanRepo     *RepoInfo   `yaml:"trainPlanRepo,omitempty"`
}

type NotificationInfo struct {
	Type string `yaml:"type"`   // "push" or "pull"
	StewardServerURI string `yaml:"stewardServerURI,omitempty"`
	PullPeriod int `yaml:"pullPeriod,omitempty"`
}

type RepoInfo struct {
	GitHttpURL string `yaml:"gitHttpURL"`
}
