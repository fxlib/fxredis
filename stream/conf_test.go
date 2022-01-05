package stream

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsumerConf(t *testing.T) {
	os.Setenv("GROUP_NAME", "foo")
	defer os.Unsetenv("GROUP_NAME")
	os.Setenv("BAR_GROUP_NAME", "bar")
	defer os.Unsetenv("BAR_GROUP_NAME")

	cfg1, err := ParseConsumerEnv("BAR_")
	require.NoError(t, err)
	require.Equal(t, "bar", cfg1.GroupName)
	cfg2, err := ParseConsumerEnv()
	require.NoError(t, err)
	require.Equal(t, "foo", cfg2.GroupName)
}
