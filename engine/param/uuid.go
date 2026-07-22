package param

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// resolveUUIDVariables resolves the uuid() function-style helper.
//
//	uuid()   -> UUID v7 (time-ordered, RFC 9562) — the default
//	uuid(7)  -> UUID v7
//	uuid(4)  -> UUID v4 (random)
//
// A fresh value is generated for every occurrence, so uuid() uuid() in the same
// command yields two distinct identifiers.
func resolveUUIDVariables(instruction string) (string, error) {
	return resolveFunction(instruction, `\buuid\(([^)]*)\)`, func(groups []string) (string, error) {
		switch strings.TrimSpace(groups[0]) {
		case "", "7":
			return uuidV7()
		case "4":
			return uuidV4()
		default:
			return "uuid(" + groups[0] + ")", nil // unknown version — leave literal
		}
	})
}

func uuidV7() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	ms := time.Now().UnixMilli()
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)

	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return formatUUID(b), nil
}

func uuidV4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return formatUUID(b), nil
}

func formatUUID(b [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
