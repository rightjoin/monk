package monk

import (
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rightjoin/rutl/conv"
)

var Monk = "123"

// NewUUID presently ignores len input
func NewUUID(length int, prefix ...string) string {

	p := ""
	if len(prefix) != 0 {
		p = strings.Join(prefix, "-") + "-"
	}

	return p + uuid.NewString()
}

var randStringSrc = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandStringBytesMaskImprSrcSB(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randStringSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randStringSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func FieldKey(rf reflect.StructField) string {
	return conv.CaseSnake(rf.Name)
}
