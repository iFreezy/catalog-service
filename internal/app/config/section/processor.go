package section

type Processor struct {
	Grpc    ProcessorGrpc
	Gateway ProcessorGateway
}

type ProcessorGrpc struct {
	ListenPort uint32 `envconfig:"APP_PROCESSOR_GRPC_LISTEN_PORT" default:"50051"`
}

type ProcessorGateway struct {
	ListenPort uint32 `envconfig:"APP_PROCESSOR_GATEWAY_LISTEN_PORT" default:"8081"`
}
