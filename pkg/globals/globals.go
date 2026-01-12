package globals

import (
	"log/slog"
	"sync"
)

var (
	mu         sync.RWMutex
	logger     *slog.Logger
	debugLevel int
)

// Functions for setting and getting global variables
func SetDebugLevel(x int) {
	mu.Lock()
	debugLevel = x
	mu.Unlock()
}

func GetDebugLevel() (x int) {
	mu.Lock()
	x = debugLevel
	mu.Unlock()
	return x
}

func SetLogger(x *slog.Logger) {
	mu.Lock()
	logger = x
	mu.Unlock()
}

func GetLogger() (x *slog.Logger) {
	mu.Lock()
	x = logger
	mu.Unlock()
	return x
}
