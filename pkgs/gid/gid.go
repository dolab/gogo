package gid

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"sync/atomic"
	"time"
)

// NOTE: copied from http://bazaar.launchpad.net/+branch/mgo/v2/view/head:/bson/bson.go#L161

var (
	// macID stores mac addr id generated once and used in subsequent calls
	// to New function.
	macID = func() []byte {
		var mac [3]byte

		ifaces, err := net.Interfaces()
		if err == nil {
			for _, iface := range ifaces {
				if iface.Flags&net.FlagUp != 0 && bytes.Compare(iface.HardwareAddr, nil) != 0 {
					// Don't use random as we have a real address
					md5hash := md5.New()
					md5hash.Write(iface.HardwareAddr)
					copy(mac[:], md5hash.Sum(nil))

					return mac[:]
				}
			}
		}

		// otherwise, use a rand string
		_, err = io.ReadFull(rand.Reader, mac[:])
		if err != nil {
			panic("cannot grab an random string: " + err.Error())
		}

		return mac[:]
	}()

	// gidCounter is atomically incremented when generating a new GID
	// using New() function. It's used as a counter part of an id.
	gidCounter uint32
)

// GID is a sortable unique ID identifying a value.
// NOTE: It must be exactly 12 bytes long.
type GID []byte

// New returns a new unique GID.
func New() GID {
	var id [12]byte

	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(id[:], uint32(time.Now().Unix()))

	// Machine, first 3 bytes of md5(hostname)
	id[4] = macID[0]
	id[5] = macID[1]
	id[6] = macID[2]

	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	id[7] = byte(pid >> 8)
	id[8] = byte(pid)

	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&gidCounter, 1)
	id[9] = byte(i >> 16)
	id[10] = byte(i >> 8)
	id[11] = byte(i)

	return GID(id[:])
}

// NewWithTime returns a dummy GID with the timestamp part filled
// with the provided number of seconds from epoch UTC, and all other parts
// filled with zeroes. It is useful only for queries filter with ids
// generated before or after the specified timestamp.
func NewWithTime(t time.Time) GID {
	var id [12]byte

	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(id[:4], uint32(t.Unix()))

	return GID(id[:])
}

// Valid returns true if id is valid. A valid id must contain exactly 12 bytes.
func (id GID) Valid() bool {
	return len(id) == 12
}

// Hex returns a hex representation of the GID.
func (id GID) Hex() string {
	return hex.EncodeToString(id)
}

// Time returns the timestamp part of the id.
// It's a runtime error to call this method with an invalid id.
func (id GID) Time() time.Time {
	// First 4 bytes of GID is 32-bit big-endian seconds from epoch.
	secs := int64(binary.BigEndian.Uint32(id.slice(0, 4)))

	return time.Unix(secs, 0)
}

// Mac returns the 3-byte mac addr id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id GID) Mac() []byte {
	return id.slice(4, 7)
}

// Pid returns the process id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id GID) Pid() uint16 {
	return binary.BigEndian.Uint16(id.slice(7, 9))
}

// Counter returns the incrementing value part of the id.
// It's a runtime error to call this method with an invalid id.
func (id GID) Counter() int32 {
	b := id.slice(9, 12)

	// Counter is stored as big-endian 3-byte value
	return int32(uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]))
}

// String returns a hex string representation of the id.
// Example: Hex("4d88e15b60f486e428412dc9").
func (id GID) String() string {
	return fmt.Sprintf(`Hex("%x")`, string(id))
}

// slice returns byte slice of id from start to end.
// Calling this function with an invalid id will cause a runtime panic.
func (id GID) slice(start, end int) []byte {
	if len(id) != 12 {
		panic(fmt.Sprintf("Invalid GID: %q", string(id)))
	}

	return id[start:end]
}

// IsHex returns whether s is a valid hex representation of
// a GID. See the Hex function.
func IsHex(s string) bool {
	if len(s) != 24 {
		return false
	}

	_, err := hex.DecodeString(s)
	return err == nil
}

// Hex returns a GID from the provided hex representation.
// Calling this function with an invalid hex representation will
// cause a runtime panic. See the IsHex function.
func Hex(s string) GID {
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 12 {
		panic(fmt.Sprintf("Invalid GID Hex: %q", s))
	}

	return GID(b)
}
