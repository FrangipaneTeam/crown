package tracker

import (
	"encoding/json"
	"fmt"
	"time"
)

type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes an int64 timestamp into a time.Time object
func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	// 1. Decode the bytes into an int64
	var raw int64
	err := json.Unmarshal(bytes, &raw)

	if err != nil {
		fmt.Printf("error decoding timestamp: %s\n", err)
		return err
	}

	// 2. Parse the unix timestamp
	p.Time = time.Unix(raw, 0)
	return nil
}

// MarshalJSON encodes a time.Time object into an int64 timestamp
func (p Timestamp) MarshalJSON() ([]byte, error) {
	// 1. Convert the time.Time object into a unix timestamp
	timestamp := p.Unix()

	// 2. Encode the unix timestamp as a JSON number
	return json.Marshal(timestamp)
}

// Now returns the current time
func (p *Timestamp) Now() {
	p.Time = time.Now()
}
