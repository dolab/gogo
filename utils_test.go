package gogo

import (
	"testing"
	"time"

	"github.com/golib/assert"
)

func Test_NewObjectId(t *testing.T) {
	assertion := assert.New(t)

	// Generate 10 ids
	ids := make([]ObjectId, 10)
	for i := 0; i < 10; i++ {
		ids[i] = NewObjectId()
	}

	for i := 1; i < 10; i++ {
		id := ids[i]
		prevId := ids[i-1]

		// Test for uniqueness among all other 9 generated ids
		for j, tid := range ids {
			if j != i {
				assertion.NotEqual(id, tid)
			}
		}

		// Check that timestamp was incremented and is within 30 seconds of the previous one
		assertion.InDelta(prevId.Time().Second(), id.Time().Second(), 0.1)

		// Check that machine ids are the same
		assertion.Equal(prevId.Machine(), id.Machine())

		// Check that pids are the same
		assertion.Equal(prevId.Pid(), id.Pid())

		// Test for proper increment
		assertion.Equal(1, int(id.Counter()-prevId.Counter()))
	}
}

func Test_NewObjectIdWithTime(t *testing.T) {
	ts := time.Unix(12345678, 0)
	id := NewObjectIdWithTime(ts)
	assertion := assert.New(t)
	assertion.Equal(ts, id.Time())
	assertion.Equal([]byte{0x00, 0x00, 0x00}, id.Machine())
	assertion.EqualValues(0, id.Pid())
	assertion.EqualValues(0, id.Counter())
}

func Test_IsObjectIdHex(t *testing.T) {
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
		assertion.Equal(testCase.valid, IsObjectIdHex(testCase.id))
	}
}

func Test_ObjectIdHex(t *testing.T) {
	s := "4d88e15b60f486e428412dc9"
	id := ObjectIdHex(s)
	assertion := assert.New(t)
	assertion.True(id.Valid())
	assertion.Equal(s, id.Hex())
	assertion.EqualValues(1300816219, id.Time().Unix())
	assertion.EqualValues(58408, id.Pid())
	assertion.Equal([]byte{0x60, 0xf4, 0x86}, id.Machine())
	assertion.EqualValues(4271561, id.Counter())
}
