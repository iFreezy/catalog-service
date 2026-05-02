package section

type Monitor struct {
	LogLevel    string `envconfig:"APP_LOG_LEVEL"   default:"debug"`
	Environment string `envconfig:"APP_ENVIRONMENT" default:"dev"`
}
