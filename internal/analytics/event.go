package analytics

const (
	eventRun             = "xbom_command_run"
	eventCommandGenerate = "xbom_command_generate"

	eventXbomGenerateEnvDocker        = "xbom_command_generate_env_docker"
	eventXbomGenerateEnvGitHubActions = "xbom_command_generate_env_github_actions"
	eventXbomGenerateEnvGitLabCI      = "xbom_command_generate_env_gitlab_ci"
)

func TrackCommandRun() {
	TrackEvent(eventRun)
}

func TrackCommandGenerate() {
	TrackEvent(eventCommandGenerate)
}

func TrackCommandGenerateEnvDocker() {
	TrackEvent(eventXbomGenerateEnvDocker)
}

func TrackCommandGenerateEnvGitHubActions() {
	TrackEvent(eventXbomGenerateEnvGitHubActions)
}

func TrackCommandGenerateEnvGitLabCI() {
	TrackEvent(eventXbomGenerateEnvGitLabCI)
}
