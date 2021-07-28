package task

import "time"

type Minutes struct {
	time.Duration
}

func NewMinutes(m int) *Minutes {
	return &Minutes{Duration: minutesDuration(m)}
}

func minutesDuration(m int) time.Duration {
	return time.Minute * time.Duration(m)
}

func (m *Minutes) MarshalYAML() (interface{}, error) {
	return int(m.Minutes()), nil
}

func (m *Minutes) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var interval int
	if err := unmarshal(&interval); err != nil {
		return err
	}
	m.Duration = minutesDuration(interval)

	return nil
}
