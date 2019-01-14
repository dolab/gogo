package gogo

import (
	"testing"
	"time"

	"github.com/golib/assert"
)

func Test_NewGID(t *testing.T) {
	it := assert.New(t)

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
				it.NotEqual(id, tid)
			}
		}

		// Check that timestamp was incremented and is within 30 seconds of the previous one
		it.InDelta(prevID.Time().Second(), id.Time().Second(), 0.1)

		// Check that mac ids are the same
		it.Equal(prevID.Mac(), id.Mac())

		// Check that pids are the same
		it.Equal(prevID.Pid(), id.Pid())

		// Test for proper increment
		it.Equal(1, int(id.Counter()-prevID.Counter()))
	}
}

func Benchmark_NewGID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewGID()
	}
}

func Test_NewGIDWithTime(t *testing.T) {
	ts := time.Unix(12345678, 0)
	id := NewGIDWithTime(ts)

	it := assert.New(t)
	it.Equal(ts, id.Time())
	it.Equal([]byte{0x00, 0x00, 0x00}, id.Mac())
	it.EqualValues(0, id.Pid())
	it.EqualValues(0, id.Counter())
}

func Test_IsGIDHex(t *testing.T) {
	it := assert.New(t)

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
		it.Equal(testCase.valid, IsGIDHex(testCase.id))
	}
}

func Test_GIDHex(t *testing.T) {
	s := "4d88e15b60f486e428412dc9"
	id := GIDHex(s)

	it := assert.New(t)
	it.True(id.Valid())
	it.Equal(s, id.Hex())
	it.EqualValues(1300816219, id.Time().Unix())
	it.EqualValues(58408, id.Pid())
	it.Equal([]byte{0x60, 0xf4, 0x86}, id.Mac())
	it.EqualValues(4271561, id.Counter())
}
