package types

import (
	"bytes"
	"strconv"
	"time"
)

const DiscordEpoch = 1420070400000

type Snowflake uint64

func ParseSnowflake(s string) (Snowflake, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return Snowflake(v), nil
}

func (s Snowflake) String() string {
	return strconv.FormatUint(uint64(s), 10)
}

func (s Snowflake) Time() time.Time {
	if s == 0 {
		return time.Time{}
	}
	msec := int64((s >> 22) + DiscordEpoch)
	return time.UnixMilli(msec)
}

func (s Snowflake) IsValid() bool {
	return s > 0
}

func (s Snowflake) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 22)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(s), 10)
	return append(buf, '"'), nil
}

func (s *Snowflake) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, `"`)
	if len(b) == 0 || bytes.Equal(b, []byte("null")) {
		*s = 0
		return nil
	}
	v, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return err
	}
	*s = Snowflake(v)
	return nil
}

