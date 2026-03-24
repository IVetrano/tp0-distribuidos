package common

import (
	"time"
	"os"
	"strconv"
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
	QuerySleepMillis int
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

func (c *Client) sendBets(signalChannel chan os.Signal, protocol *Protocol, csvIterator *CSVBetIterator) error {
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
			return err
		}

		if bets == nil {
			break
		}

		err = protocol.SendBetBatch(bets)
		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return 	err
		}
	}
	return nil
}

func (c *Client) getWinners() ([]string, error) {
	for {
		protocol, err := NewProtocol(c.config.ServerAddress)
		if err != nil {
			log.Errorf("action: connect_query | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return nil, err
		}
		
		agencyID, err := strconv.Atoi(c.config.ID)
		if err != nil {
			log.Errorf("action: parse_agency_id | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return nil, err
		}

		ready, winners, err := protocol.QueryWinners(agencyID)
		if err != nil {
			log.Errorf("action: query_winners | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return nil, err
		}

		err = protocol.Close()
		if err != nil {
			log.Errorf("action: close_connection_query | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return nil, err
		}

		if ready {
			return winners, nil
		}

		time.Sleep(time.Duration(c.config.QuerySleepMillis) * time.Millisecond)
	}
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
	
	// Send bets
	err = c.sendBets(signalChannel, protocol, csvIterator)
	if err != nil {
		return
	}
	log.Infof("action: send_bets | result: success")

	// Close the CSV file
	err = csvIterator.Close()
	if err != nil {
		log.Errorf("action: close_csv_iterator | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	// Send finish
	err = protocol.SendFinish()
	if err != nil {
		log.Errorf("action: send_finish | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	log.Infof("action: send_finish | result: success")

	// Close the connection
	err = protocol.Close()
	if err != nil {
		log.Errorf("action: close_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	// Get winners
	winners, err := c.getWinners()
	if err != nil {
		return
	}
	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
