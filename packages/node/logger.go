package node

import (
	"fmt"
	"sync"
)

type Logger struct {
	Enabled    bool
	LogInfo    func(pluginName string, message string)
	LogSuccess func(pluginName string, message string)
	LogWarning func(pluginName string, message string)
	LogFailure func(pluginName string, message string)
	LogDebug   func(pluginName string, message string)
	mutex      sync.RWMutex
}

func (logger *Logger) SetEnabled(value bool) {
	logger.mutex.Lock()
	logger.Enabled = value
	logger.mutex.Unlock()
}

func (logger *Logger) GetEnabled() bool {
	logger.mutex.RLock()
	defer logger.mutex.RUnlock()
	return logger.Enabled
}

func pluginPrefix(pluginName string) string {
	var pluginPrefix string
	if pluginName == "Node" {
		pluginPrefix = ""
	} else {
		pluginPrefix = pluginName + ": "
	}

	return pluginPrefix
}

var DEFAULT_LOGGER = &Logger{
	Enabled: true,
	LogSuccess: func(pluginName string, message string) {
		fmt.Println("[  OK  ] " + pluginPrefix(pluginName) + message)
	},
	LogInfo: func(pluginName string, message string) {
		fmt.Println("[ INFO ] " + pluginPrefix(pluginName) + message)
	},
	LogWarning: func(pluginName string, message string) {
		fmt.Println("[ WARN ] " + pluginPrefix(pluginName) + message)
	},
	LogFailure: func(pluginName string, message string) {
		fmt.Println("[ FAIL ] " + pluginPrefix(pluginName) + message)
	},
	LogDebug: func(pluginName string, message string) {
		fmt.Println("[ NOTE ] " + pluginPrefix(pluginName) + message)
	},
}
