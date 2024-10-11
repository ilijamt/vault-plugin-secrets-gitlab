//go:build !integration

package gitlab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToInt(t *testing.T) {
	var tests = []struct {
		in     any
		outVal int
		outErr error
	}{
		{int(52), int(52), nil},
		{int8(13), int(13), nil},
		{int16(612), int(612), nil},
		{int32(56236), int(56236), nil},
		{int64(23462346), int(23462346), nil},
		{float32(62346.62), int(62346), nil},
		{float64(263467.26), int(263467), nil},
		{"1", int(0), ErrInvalidValue},
	}

	for _, tst := range tests {
		t.Logf("convertToInt(%T(%v))", tst.in, tst.in)
		val, err := convertToInt(tst.in)
		assert.Equal(t, tst.outVal, val)
		if tst.outErr != nil {
			assert.ErrorIs(t, err, tst.outErr)
		}
	}
}

func TestCalculateGitlabTTL(t *testing.T) {
	locMST, err := time.LoadLocation("MST")
	require.NoError(t, err)
	var tests = []struct {
		inDuration  time.Duration
		inTime      time.Time
		outDuration time.Duration
		outExpiry   time.Time
		outErr      error
	}{
		// 1h on 2024-02-22T13:06:10.575Z, should expire 2024-02-23
		{
			inDuration:  time.Hour,
			inTime:      time.Date(2024, 2, 22, 13, 6, 10, 0, time.UTC),
			outDuration: (time.Hour * 10) + (53 * time.Minute) + (50 * time.Second),
			outExpiry:   time.Date(2024, 2, 23, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},
		// 3h
		{
			inDuration:  3 * time.Hour,
			inTime:      time.Date(2024, 2, 22, 13, 0, 0, 0, time.UTC),
			outDuration: time.Hour * 11,
			outExpiry:   time.Date(2024, 2, 23, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		// 1h1s
		{
			inDuration:  time.Hour + time.Second,
			inTime:      time.Date(2024, 2, 22, 23, 0, 0, 0, time.UTC),
			outDuration: time.Hour * 25,
			outExpiry:   time.Date(2024, 2, 24, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		// 23h on 2024-02-22T20:00:00.000Z, should expire 2024-02-24
		{
			inDuration:  time.Hour * 23,
			inTime:      time.Date(2024, 2, 22, 20, 0, 0, 0, time.UTC),
			outDuration: 28 * time.Hour,
			outExpiry:   time.Date(2024, 2, 24, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		// 5h on 2024-02-22T20:00:00.000Z, should expire 2024-02-24
		{
			inDuration:  time.Hour * 5,
			inTime:      time.Date(2024, 2, 22, 20, 0, 0, 0, time.UTC),
			outDuration: time.Hour * 28,
			outExpiry:   time.Date(2024, 2, 24, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		// 45h on 2024-02-22T20:00:00.000Z, should expire 2024-02-25
		{
			inDuration:  time.Hour * 45,
			inTime:      time.Date(2024, 2, 22, 20, 0, 0, 0, time.UTC),
			outDuration: time.Hour * 52,
			outExpiry:   time.Date(2024, 2, 25, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		{
			inDuration: time.Hour * 2,
			// 2024-05-30 13:01:43 -0600 MDT
			inTime:      time.Date(2024, 5, 30, 13, 01, 43, 0, locMST),
			outDuration: time.Hour*3 + time.Minute*58 + time.Second*17,
			outExpiry:   time.Date(2024, 5, 31, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},

		{
			inDuration:  390 * 25 * time.Hour,
			inTime:      time.Date(2024, 7, 11, 15, 41, 0, 0, time.UTC),
			outDuration: 8744*time.Hour + 19*time.Minute,
			outExpiry:   time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC),
			outErr:      nil,
		},
	}

	for _, tst := range tests {
		t.Logf("calculateGitlabTTL(%s, %s) = duration %s, expiry %s, error %v", tst.inDuration, tst.inTime.Format(time.RFC3339), tst.outDuration, tst.outExpiry.Format(time.RFC3339), tst.outErr)
		dur, exp, err := calculateGitlabTTL(tst.inDuration, tst.inTime)
		if err != nil {
			assert.ErrorIs(t, err, tst.outErr)
		}
		assert.EqualValues(t, tst.outExpiry, exp)
		assert.WithinDuration(t, tst.outExpiry, exp, time.Minute)
		assert.EqualValues(t, tst.outDuration, dur)
	}
}
