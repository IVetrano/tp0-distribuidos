package common

import (
	"time"
	"os"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}



// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(signalChannel chan os.Signal, bet *Bet) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	running := true

	for msgID := 1; msgID <= c.config.LoopAmount && running; msgID++ {
		// Check if a SIGTERM signal has been received. If so, break the loop
		select {
        case <-signalChannel:
            log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
            running = false
            continue
        default:
        }

		// Create the connection the server in every loop iteration.
		protocol, err := NewProtocol(c.config.ServerAddress)
		if err != nil {
			log.Criticalf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		
		// Send the Bet
		err = protocol.SendBet(bet)
		if err != nil {
			log.Errorf("action: send_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.Document, bet.Number)

		// Close the connection
		err = protocol.Close()
		if err != nil {
			log.Errorf("action: close_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
	}

	if running {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	}
}
