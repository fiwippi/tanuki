package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestJob_MarshalYAML(t *testing.T) {
	j := NewJob(10)
	data, err := yaml.Marshal(j)
	require.Nil(t, err)
	require.Equal(t, "10\n", string(data))
}

func TestJob_UnmarshalYAML(t *testing.T) {
	var j Job
	err := yaml.Unmarshal([]byte("10"), &j)
	require.Nil(t, err)
	require.Equal(t, 10*time.Minute, j.Duration)
}
