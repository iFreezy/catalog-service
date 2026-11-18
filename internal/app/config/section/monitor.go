package section

type Monitor struct {
	LogLevel    string `envconfig:"APP_MONITOR_LOG_LEVEL"   default:"debug"`
	Environment string `envconfig:"APP_MONITOR_ENVIRONMENT" default:"dev"`
}
