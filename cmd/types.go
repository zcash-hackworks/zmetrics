package cmd

// Options provides any flags used by zmetrics
type Options struct {
	LogLevel    uint64 `json:"log_level,omitempty"`
	RPCUser     string `json:"rpcUser,omitempty"`
	RPCPassword string `json:"rpcPassword,omitempty"`
	RPCHost     string `json:"rpcHost,omitempty"`
	RPCPort     string `json:"rpcPort,omitempty"`
}
