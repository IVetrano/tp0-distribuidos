package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type CSVBetIterator struct {
	file   *os.File
	reader *bufio.Reader
	agencyID string
	done  bool
}

func NewCSVBetIterator(filePath string, agencyID string) (*CSVBetIterator, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %v", err)
	}

	return &CSVBetIterator{
		file:   file,
		reader: bufio.NewReader(file),
		agencyID: agencyID,
	}, nil
}

func (it *CSVBetIterator) Done() bool {
	return it.done
}

func (it *CSVBetIterator) NextBatch(maxAmount int) ([]*Bet, error) {
    if it.done {
        return nil, nil
    }

    bets := make([]*Bet, 0, maxAmount)

	// Read lines from the csv until there are maxAmount bets or is the end of the file
    for len(bets) < maxAmount {
        line, err := it.reader.ReadString('\n')
        if err != nil && err != io.EOF {
            return nil, fmt.Errorf("error reading CSV line: %v", err)
        }

        line = strings.TrimSpace(line)
        if line != "" {
            fields := strings.Split(line, ",")
            if len(fields) != 5 {
                it.done = true
                return nil, fmt.Errorf("invalid CSV format: expected 5 fields per line")
            }

            bet, err := NewBetFromStrings(
                it.agencyID,
                fields[0],
                fields[1],
                fields[2],
                fields[3],
                fields[4],
            )
            if err != nil {
                it.done = true
                return nil, fmt.Errorf("error parsing bet from CSV record: %v", err)
            }
            bets = append(bets, bet)
        }

        if err == io.EOF {
            it.done = true
            break
        }
    }

    if len(bets) == 0 && it.done {
        return nil, nil
    }

    return bets, nil
}

func (it *CSVBetIterator) Close() error {
	err := it.file.Close()
	if err != nil {
		return fmt.Errorf("error closing CSV file: %v", err)
	}
	return nil
}
