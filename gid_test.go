package gogo

import (
	"testing"
	"time"

	"github.com/golib/assert"
)

func Test_NewGID(t *testing.T) {
	assertion := assert.New(t)

	// Generate 10 ids
	ids := make([]GID, 10)
	for i := 0; i < 10; i++ {
		ids[i] = NewGID()
	}

	for i := 1; i < 10; i++ {
		id := ids[i]
		prevID := ids[i-1]

		// Test for uniqueness among all other 9 generated ids
		for j, tid := range ids {
			if j != i {
				assertion.NotEqual(id, tid)
			}
		}

		// Check that timestamp was incremented and is within 30 seconds of the previous one
		assertion.InDelta(prevID.Time().Second(), id.Time().Second(), 0.1)

		// Check that mac ids are the same
		assertion.Equal(prevID.Mac(), id.Mac())

		// Check that pids are the same
		assertion.Equal(prevID.Pid(), id.Pid())

		// Test for proper increment
		assertion.Equal(1, int(id.Counter()-prevID.Counter()))
	}
}

func Test_NewGIDWithTime(t *testing.T) {
	ts := time.Unix(12345678, 0)
	id := NewGIDWithTime(ts)

	assertion := assert.New(t)
	assertion.Equal(ts, id.Time())
	assertion.Equal([]byte{0x00, 0x00, 0x00}, id.Mac())
	assertion.EqualValues(0, id.Pid())
	assertion.EqualValues(0, id.Counter())
}

func Test_IsGIDHex(t *testing.T) {
	assertion := assert.New(t)

	testCases := []struct {
		id    string
		valid bool
	}{
		{"4d88e15b60f486e428412dc9", true},
		{"4d88e15b60f486e428412dc", false},
		{"4d88e15b60f486e428412dc9e", false},
		{"4d88e15b60f486e428412dcx", false},
	}

	for _, testCase := range testCases {
		assertion.Equal(testCase.valid, IsGIDHex(testCase.id))
	}
}

func Test_GIDHex(t *testing.T) {
	s := "4d88e15b60f486e428412dc9"
	id := GIDHex(s)

	assertion := assert.New(t)
	assertion.True(id.Valid())
	assertion.Equal(s, id.Hex())
	assertion.EqualValues(1300816219, id.Time().Unix())
	assertion.EqualValues(58408, id.Pid())
	assertion.Equal([]byte{0x60, 0xf4, 0x86}, id.Mac())
	assertion.EqualValues(4271561, id.Counter())
}
