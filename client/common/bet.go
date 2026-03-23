package common

import (
	"fmt"
	"os"
	"time"
	"strconv"
)

type Bet struct {
	Agency int
	FirstName string
	LastName string
	Document string
	BirthDate time.Time
	Number int
}

func NewBetFromStrings(agencyID, firstName, lastName, document, birthDateStr, numberStr string) (*Bet, error) {
    if firstName == "" || lastName == "" || document == "" {
        return nil, fmt.Errorf("missing required fields")
    }

    agency, err := strconv.Atoi(agencyID)
    if err != nil {
        return nil, fmt.Errorf("error parsing agency ID: %v", err)
    }

    birthDate, err := time.Parse("2006-01-02", birthDateStr)
    if err != nil {
        return nil, fmt.Errorf("error parsing birth date: %v", err)
    }

    number, err := strconv.Atoi(numberStr)
    if err != nil {
        return nil, fmt.Errorf("error parsing number: %v", err)
    }

    return &Bet{
        Agency:    agency,
        FirstName: firstName,
        LastName:  lastName,
        Document:  document,
        BirthDate: birthDate,
        Number:    number,
    }, nil
}
