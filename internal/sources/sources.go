// internal/sources/source.go
package sources

import (
	"github.com/Yriskit-ai/logflow/internal/ipc"
)

// Source represents a log source that can stream log entries
type Source interface {
	Stream(client *ipc.Client) error
	Name() string
	Type() string
}
