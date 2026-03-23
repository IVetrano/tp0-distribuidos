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
	CsvFilePath   string
	MaxBatchAmount int
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



func (c *Client) StartClientLoop(signalChannel chan os.Signal) {
	// Create the connection the server in every loop iteration.
	protocol, err := NewProtocol(c.config.ServerAddress)
	if err != nil {
		log.Criticalf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	// Create the CSV iterator for the current client
	csvIterator, err := NewCSVBetIterator(c.config.CsvFilePath, c.config.ID)
	if err != nil {
		log.Criticalf("action: create_csv_iterator | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	
	// Send Bets
	for !csvIterator.Done() {
		// Check if a SIGTERM signal has been received. If so, break the loop
		select {
        case <-signalChannel:
            log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
            break
        default:
        }

		bets, err := csvIterator.NextBatch(c.config.MaxBatchAmount)
		if err != nil {
			log.Errorf("action: read_csv_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		if bets == nil {
			break
		}

		err = protocol.SendBetBatch(bets)
		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
	}

	log.Infof("action: send_bets | result: success")

	// Close the CSV file
	err = csvIterator.Close()
	if err != nil {
		log.Errorf("action: close_csv_iterator | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	// Close the connection
	err = protocol.Close()
	if err != nil {
		log.Errorf("action: close_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
