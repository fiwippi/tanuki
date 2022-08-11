package task

func (j *Job) MarshalYAML() (interface{}, error) {
	return int(j.Minutes()), nil
}

func (j *Job) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var interval int
	if err := unmarshal(&interval); err != nil {
		return err
	}
	j.Duration = minutesDuration(interval)

	return nil
}
