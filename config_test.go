package config

import (
	"log"
	"os"
	"path"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
)

type testConfig struct {
	Name string
}

func setupTests(t *testing.T) string {
	dir := t.TempDir()
	file := path.Join(dir, "config.json")

	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f.WriteString(`{"Name": "Nic"}`)

	return file
}

func modifyFile(f string, data string) error {
	// delete the old file
	err := os.Remove(f)
	if err != nil {
		return xerrors.Errorf("error removing config file: %w", err)
	}

	fi, err := os.OpenFile(f, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return xerrors.Errorf("error creating file: %w", err)
	}
	defer fi.Close()

	_, err = fi.WriteString(data)
	if err != nil {
		return xerrors.Errorf("error writing update to file: %w", err)
	}

	return nil
}

func TestLoadsConfigIntoStructOnStart(t *testing.T) {
	filePath := setupTests(t)

	fw, err := New(filePath, 1*time.Millisecond, &log.Logger{}, func(*testConfig) {})
	require.NoError(t, err)
	defer fw.Close()

	tc := fw.Read()

	assert.NoError(t, err)
	assert.Equal(t, "Nic", tc.Name)
}

func TestLoadsConfigIntoStructOnChange(t *testing.T) {
	filePath := setupTests(t)

	fw, err := New[*testConfig](filePath, 1*time.Millisecond, &log.Logger{}, nil)
	require.NoError(t, err)
	defer fw.Close()

	tc := fw.Read()
	require.Equal(t, "Nic", tc.Name)

	// modify the config
	err = modifyFile(filePath, `{"Name": "Erik"}`)
	require.NoError(t, err)

	assert.Eventually(t,
		func() bool {
			tc = fw.Read()
			return tc != nil && tc.Name == "Erik"
		}, 5*time.Second, 1*time.Millisecond,
	)
}

func TestCallsUpdateOnChange(t *testing.T) {
	updated := &atomic.Bool{}
	updated.Store(false)
	filePath := setupTests(t)

	fw, err := New(filePath, 1*time.Millisecond, &log.Logger{}, func(*testConfig) { updated.Store(true) })
	require.NoError(t, err)
	defer fw.Close()

	tc := fw.Read()
	require.Equal(t, "Nic", tc.Name)

	// modify the config
	err = modifyFile(filePath, `{"Name": "Erik"}`)
	require.NoError(t, err)

	assert.Eventually(t,
		func() bool {
			return updated.Load()
		}, 5*time.Second, 1*time.Millisecond,
	)
}
