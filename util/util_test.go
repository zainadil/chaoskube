package util

import (
	"testing"
	"time"
)

func TestTimePeriodIncludes(t *testing.T) {
	for _, tc := range []struct {
		pointInTime time.Time
		timePeriod  TimePeriod
		expected    bool
	}{
		// it's included
		{
			time.Now(),
			TimePeriod{From: time.Now().Add(-1 * time.Minute), To: time.Now().Add(+1 * time.Minute)},
			true,
		},
		// it's before
		{
			time.Now().Add(-2 * time.Minute),
			TimePeriod{From: time.Now().Add(-1 * time.Minute), To: time.Now().Add(+1 * time.Minute)},
			false,
		},
		// it's after
		{
			time.Now().Add(+2 * time.Minute),
			TimePeriod{From: time.Now().Add(-1 * time.Minute), To: time.Now().Add(+1 * time.Minute)},
			false,
		},
	} {
		got := tc.timePeriod.Includes(tc.pointInTime)
		if tc.expected != got {
			t.Fatalf("expected %v, got %v", tc.expected, got)
		}
	}
}

func TestParseWeekdays(t *testing.T) {
	for _, tc := range []struct {
		given    string
		expected []time.Weekday
	}{
		// empty string
		{
			"",
			[]time.Weekday{},
		},
		// single weekday
		{
			"sat",
			[]time.Weekday{time.Saturday},
		},
		// multiple weekdays
		{
			"sat,sun",
			[]time.Weekday{time.Saturday, time.Sunday},
		},
		// case-insensitive
		{
			"SaT,SUn",
			[]time.Weekday{time.Saturday, time.Sunday},
		},
		// ignore whitespace
		{
			" sat , sun ",
			[]time.Weekday{time.Saturday, time.Sunday},
		},
		// ignore unknown weekdays
		{
			"sat,unknown,sun",
			[]time.Weekday{time.Saturday, time.Sunday},
		},
		// deal with all kinds at the same time
		{
			"Fri, sat ,,,,  ,foobar,tue",
			[]time.Weekday{time.Friday, time.Saturday, time.Tuesday},
		},
	} {
		got := ParseWeekdays(tc.given)

		if len(tc.expected) != len(got) {
			t.Fatalf("expected %d, got %d", len(tc.expected), len(got))
		}

		for i := range tc.expected {
			if tc.expected[i] != got[i] {
				t.Errorf("expected %v, got %v", tc.expected[i], got[i])
			}
		}
	}
}
