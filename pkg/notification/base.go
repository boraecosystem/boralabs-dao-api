package notification

import "log"

var BaseLoggers []Logger

// Logger is the interface that wraps the SendMessage method.
type Logger interface {
	SendMessage(message string) error
}

// SendAll sends a message by all loggers.
func SendAll(message string) {
	for _, logger := range BaseLoggers {
		if err := logger.SendMessage(message); err != nil {
			log.Println(err.Error())
		}
	}
}
