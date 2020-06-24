package aggregateServer

type Payload struct {
	GrpcServerURI string
	TrainPlanRepoGitHttpURL string
	AggregatedModelRepoGitHttpURL string
	EdgeModelRepoGitHttpURLs []string
}