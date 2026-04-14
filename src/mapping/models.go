package mapping

type ServiceStatus string
type DatabaseConnectionStatus string

const (
	Ok    ServiceStatus = "ok"
	Error ServiceStatus = "error"
)

const (
	Connected    DatabaseConnectionStatus = "connected"
	Disconnected DatabaseConnectionStatus = "disconnected"
)

type InfoResponse struct {
	ServiceName string   `json:"service_name"`
	Version     string   `json:"version"`
	Services    []string `json:"services"`
	Environment string   `json:"environment"`
}

type HealthResponse struct {
	Version     string                   `json:"version"`
	Environment string                   `json:"environment"`
	Database    DatabaseConnectionStatus `json:"database"`
	Status      ServiceStatus            `json:"status"`
}
