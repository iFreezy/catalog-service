package section

import "time"

type WebServer struct {
	Address      string        `envconfig:"APP_WEB_SERVER_ADDRESS"       default:":8080"`
	ReadTimeout  time.Duration `envconfig:"APP_WEB_SERVER_READ_TIMEOUT"   default:"30s"`
	WriteTimeout time.Duration `envconfig:"APP_WEB_SERVER_WRITE_TIMEOUT"  default:"30s"`
	IdleTimeout  time.Duration `envconfig:"APP_WEB_SERVER_IDLE_TIMEOUT"   default:"60s"`
}
