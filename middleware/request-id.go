package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/nxtcoder17/ivy"
)

func generateRequestID() string {
	timeBuf := new(bytes.Buffer)
	if err := binary.Write(timeBuf, binary.LittleEndian, time.Now().UnixMicro()); err != nil {
		panic(err)
	}

	randBytes := make([]byte, 8)
	rand.Read(randBytes)

	hash := sha1.Sum(append(timeBuf.Bytes(), randBytes...))
	return hex.EncodeToString(hash[:])[:8]
}

func RequestID(generatorFn ...func() string) ivy.Handler {
	gen := func() func() string {
		if len(generatorFn) > 0 {
			return generatorFn[0]
		}

		return generateRequestID
	}()

	return func(c *ivy.Context) error {
		if c.GetRequestID() == "" {
			id := gen()
			c.Logger = c.Logger.With("request_id", id)
			c.SetRequestID(id)
		}
		return c.Next()
	}
}
