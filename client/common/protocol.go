package common

import (
	"fmt"
	"bytes"
	"encoding/binary"
	"net"
)

const (
	TypeBetBatch byte = 0
	TypeFinish   byte = 1
	TypeQuery    byte = 2
)

const (
	WinnersNotReady byte = 0
	WinnersReady    byte = 1
)

const firstNameSize = 24
const lastNameSize = 20
const documentSize = 12

const ackSize = 1
const queryResponseSize = 1
const winnerCountSize = 4

// ACK constants
const (
	AckOk			byte = 0
	AckBadRequest 	byte = 1
	AckServerError	byte = 2
)

type Protocol struct {
	clientSocket net.Conn
}

// Initializes protocol and client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func NewProtocol(serverAddress string) (*Protocol, error) {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, err
	}
	return &Protocol{clientSocket: conn}, nil
}

func (p *Protocol) getFixedString(str string, size int) ([]byte, error) {
	bytes := []byte(str)

	if len(bytes) > size {
		return nil, fmt.Errorf("string '%s' exceeds fixed size of %d bytes", str, size)
	}

	b := make([]byte, size)
	copy(b, bytes)
	return b, nil
}

func (p *Protocol) serializeBet(bet *Bet) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Agency (2 bytes)
	err := binary.Write(buf, binary.BigEndian, uint16(bet.Agency))
	if err != nil {
		return nil, fmt.Errorf("error serializing agency: %v", err)
	}

	// First Name (24 bytes)
	firstNameBytes, err := p.getFixedString(bet.FirstName, firstNameSize)
	if err != nil {
		return nil, fmt.Errorf("error processing first name: %v", err)
	}
	_, err = buf.Write(firstNameBytes)
	if err != nil {
		return nil, fmt.Errorf("error serializing first name: %v", err)
	}

	// Last Name (20 bytes)
	lastNameBytes, err := p.getFixedString(bet.LastName, lastNameSize)
	if err != nil {
		return nil, fmt.Errorf("error processing last name: %v", err)
	}
	_, err = buf.Write(lastNameBytes)
	if err != nil {
		return nil, fmt.Errorf("error serializing last name: %v", err)
	}

	// Document (12 bytes)
	documentBytes, err := p.getFixedString(bet.Document, documentSize)
	if err != nil {
		return nil, fmt.Errorf("error processing document: %v", err)
	}
	_, err = buf.Write(documentBytes)
	if err != nil {
		return nil, fmt.Errorf("error serializing document: %v", err)
	}

	// Birth Date (4 bytes - u16 + u8 + u8)
	err = binary.Write(buf, binary.BigEndian, uint16(bet.BirthDate.Year()))
	if err != nil {
		return nil, fmt.Errorf("error serializing birth year: %v", err)
	}
	err = binary.Write(buf, binary.BigEndian, uint8(bet.BirthDate.Month()))
	if err != nil {
		return nil, fmt.Errorf("error serializing birth month: %v", err)
	}
	err = binary.Write(buf, binary.BigEndian, uint8(bet.BirthDate.Day()))
	if err != nil {
		return nil, fmt.Errorf("error serializing birth day: %v", err)
	}

	// Number (4 bytes)
	err = binary.Write(buf, binary.BigEndian, uint32(bet.Number))
	if err != nil {
		return nil, fmt.Errorf("error serializing number: %v", err)
	}

	return buf.Bytes(), nil
}

func (p *Protocol) sendAll(data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := p.clientSocket.Write(data[totalSent:])
		if err != nil {
			return fmt.Errorf("error sending data: %v", err)
		}
		totalSent += n
	}
	return nil
}

func (p *Protocol) receiveNBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	totalRead := 0

	for totalRead < n {
		readBytes, err := p.clientSocket.Read(buf[totalRead:])
		if err != nil {
			return nil, fmt.Errorf("error receiving data: %v", err)
		}
		totalRead += readBytes
	}

	return buf, nil
}

func (p *Protocol) receiveAck() (byte, error) {
	ack, err := p.receiveNBytes(ackSize)
	if err != nil {
		return 0, fmt.Errorf("error receiving ack: %v", err)
	}
	return ack[0], nil
}

func (p *Protocol) sendAndWaitAck(data []byte) error {
	err := p.sendAll(data)
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	ack, err := p.receiveAck()
	if err != nil {
		return fmt.Errorf("error receiving ack: %v", err)
	}

	switch ack {
	case AckOk:
		return nil
	case AckBadRequest:
		return fmt.Errorf("bad request")
	case AckServerError:
		return fmt.Errorf("server error")
	default:
		return fmt.Errorf("unknown ack value: %v", ack)
	}
}

func (p *Protocol) serializeBetBatch(bets []*Bet) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Type (1 byte)
	err := binary.Write(buf, binary.BigEndian, TypeBetBatch)
	if err != nil {
		return nil, fmt.Errorf("error serializing message type: %v", err)
	}

	// Amount (4 bytes)
	err = binary.Write(buf, binary.BigEndian, uint32(len(bets)))
	if err != nil {
		return nil, fmt.Errorf("error serializing bet count: %v", err)
	}

	for _, bet := range bets {
		serializedBet, err := p.serializeBet(bet)
		if err != nil {
			return nil, fmt.Errorf("error serializing bet: %v", err)
		}
		_, err = buf.Write(serializedBet)
		if err != nil {
			return nil, fmt.Errorf("error serializing bet data: %v", err)
		}
	}

	return buf.Bytes(), nil
}

func (p *Protocol) SendBetBatch(bets []*Bet) error {
	serializedBatch, err := p.serializeBetBatch(bets)
	if err != nil {
		return err
	}
	return p.sendAndWaitAck(serializedBatch)
}

func (p *Protocol) SendFinish() error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, TypeFinish)
	if err != nil {
		return fmt.Errorf("error serializing finish message: %v", err)
	}

	return p.sendAndWaitAck(buf.Bytes())
}

func (p *Protocol) getWinners() ([]string, error) {
    countBytes, err := p.receiveNBytes(winnerCountSize)
    if err != nil {
        return nil, fmt.Errorf("error receiving winner count: %v", err)
    }

    count := binary.BigEndian.Uint32(countBytes)
    winners := make([]string, count)

    for i := uint32(0); i < count; i++ {
        winnerBytes, err := p.receiveNBytes(documentSize)
        if err != nil {
            return nil, fmt.Errorf("error receiving winner document: %v", err)
        }

		winner := string(bytes.TrimRight(winnerBytes, "\x00"))
        winners[i] = winner
    }

    return winners, nil
}

func (p *Protocol) QueryWinners(agencyID int) (ready bool, winners []string, err error) {
	buf := new(bytes.Buffer)

	// Type (1 byte)
	err = binary.Write(buf, binary.BigEndian, TypeQuery)
	if err != nil {
		return false, nil, fmt.Errorf("error serializing query message: %v", err)
	}

	// Agency ID (2 bytes)
	err = binary.Write(buf, binary.BigEndian, uint16(agencyID))
	if err != nil {
		return false, nil, fmt.Errorf("error serializing agency ID: %v", err)
	}

	// Send query
	err = p.sendAll(buf.Bytes())
	if err != nil {
		return false, nil, fmt.Errorf("error sending query message: %v", err)
	}

	// Wait for response
	response, err := p.receiveNBytes(queryResponseSize)
	if err != nil {
		return false, nil, fmt.Errorf("error receiving query response: %v", err)
	}

	switch response[0] {
	case WinnersNotReady:
		return false, nil, nil
	case WinnersReady:
		winners, err = p.getWinners()
		if err != nil {
			return false, nil, fmt.Errorf("error getting winners: %v", err)
		}
		return true, winners, nil
	default:
		return false, nil, fmt.Errorf("unknown query response value: %v", response[0])
	}
}

func (p *Protocol) Close() error {
	err := p.clientSocket.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %v", err)
	}
	return nil
}