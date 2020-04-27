package config

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/xerrors"
)

type testConfig struct {
	Name string
}

func setupTests(t *testing.T) string {
	os.Remove("/tmp/config.json")
	f, err := os.OpenFile("/tmp/config.json", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	f.WriteString(`{"Name": "Nic"}`)

	return "/tmp/config.json"
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

	tc := &testConfig{}
	fw, err := New(filePath, tc, &log.Logger{}, nil)
	defer fw.Close()

	assert.NoError(t, err)
	assert.Equal(t, "Nic", tc.Name)
}

func TestLoadsConfigIntoStructOnChange(t *testing.T) {
	filePath := setupTests(t)

	tc := &testConfig{}
	fw, err := New(filePath, tc, nil)
	defer fw.Close()

	// modify the config
	err = modifyFile(filePath, `{"Name": "Erik"}`)
	assert.NoError(t, err)

	assert.Eventually(t,
		func() bool {
			return tc != nil && tc.Name == "Erik"
		}, 10*time.Second, 10*time.Millisecond,
	)
}

func TestCallsUpdateOnChange(t *testing.T) {
	var updated = false
	filePath := setupTests(t)

	tc := &testConfig{}
	fw, err := New(filePath, tc, func() { updated = true })
	defer fw.Close()

	// modify the config
	err = modifyFile(filePath, `{"Name": "Erik"}`)
	assert.NoError(t, err)

	assert.Eventually(t,
		func() bool {
			return updated
		}, 10*time.Second, 10*time.Millisecond,
	)
}
