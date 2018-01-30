package util

import (
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type TimePeriod struct {
	From time.Time
	To   time.Time
}

func (tp TimePeriod) Includes(time time.Time) bool {
	return time.After(tp.From) && time.Before(tp.To)
}

// NewPod returns a new pod instance for testing purposes.
func NewPod(namespace, name string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"app": name,
			},
			Annotations: map[string]string{
				"chaos": name,
			},
		},
	}
}

// ParseWeekdays takes a comma-separated list of abbreviated weekdays (e.g. sat,sun) and turns them
// into a slice of time.Weekday. It ignores any whitespace and any invalid weekdays.
func ParseWeekdays(weekdays string) []time.Weekday {
	var days = map[string]time.Weekday{
		"sun": time.Sunday,
		"mon": time.Monday,
		"tue": time.Tuesday,
		"wed": time.Wednesday,
		"thu": time.Thursday,
		"fri": time.Friday,
		"sat": time.Saturday,
	}

	parsedWeekdays := []time.Weekday{}
	for _, wd := range strings.Split(weekdays, ",") {
		if day, ok := days[strings.TrimSpace(strings.ToLower(wd))]; ok {
			parsedWeekdays = append(parsedWeekdays, day)
		}
	}
	return parsedWeekdays
}
