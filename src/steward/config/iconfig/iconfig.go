package iconfig

type IConfig struct {
	Type                  string `yaml:"type"`
	StewardServerURI      string `yaml:"stewardServerURI"`
	OperatorGrpcServerURI string `yaml:"operatorGrpcServerURI,omitempty"`
	AppGrpcServerURI      string `yaml:"appGrpcServerURI,omitempty"`
	LogLevel              string `yaml:"logLevel"`
	LogPath               string `yaml:"logPath,omitempty"`

	// Git Associate
	GitUserToken        string      `yaml:"gitUserToken"`
	AggregatedModelRepo *RepoInfo   `yaml:"aggregatorModelRepo,omitempty"`
	EdgeModelRepos      []*RepoInfo `yaml:"edgeModelRepos,omitempty"`
	EdgeModelRepo       *RepoInfo   `yaml:"edgeModelRepo,omitempty"`
	TrainPlanRepo     *RepoInfo   `yaml:"trainPlanRepo,omitempty"`
}

type RepoInfo struct {
	GitHttpURL string `yaml:"gitHttpURL"`
}
