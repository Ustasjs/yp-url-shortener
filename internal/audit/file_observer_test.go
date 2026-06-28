package audit_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"Ustasjs/yp-url-shortener/internal/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileObserver_AppendsJSONLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	observer, err := audit.NewFileObserver(path)
	require.NoError(t, err)
	defer observer.Close()

	events := []audit.Event{
		{Timestamp: 100, Action: audit.ActionShorten, UserID: "u1", URL: "https://a"},
		{Timestamp: 200, Action: audit.ActionFollow, UserID: "u2", URL: "https://b"},
	}
	for _, ev := range events {
		observer.Notify(ev)
	}

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	var got []audit.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var ev audit.Event
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &ev))
		got = append(got, ev)
	}
	require.NoError(t, scanner.Err())

	assert.Equal(t, events, got)
}

func TestFileObserver_AppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	require.NoError(t, os.WriteFile(path, []byte("preexisting\n"), 0644))

	observer, err := audit.NewFileObserver(path)
	require.NoError(t, err)
	defer observer.Close()

	observer.Notify(audit.Event{Timestamp: 1, Action: audit.ActionShorten, URL: "https://x"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	assert.Equal(t, "preexisting", lines[0])
	assert.Len(t, lines, 2)

	var ev audit.Event
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &ev))
	assert.Equal(t, audit.ActionShorten, ev.Action)
	assert.Equal(t, "https://x", ev.URL)
}

func TestFileObserver_OmitsEmptyUserID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	observer, err := audit.NewFileObserver(path)
	require.NoError(t, err)
	defer observer.Close()

	observer.Notify(audit.Event{Timestamp: 1, Action: audit.ActionFollow, URL: "https://x"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "user_id")
}

func TestFileObserver_NewFileObserver_BadPath(t *testing.T) {
	_, err := audit.NewFileObserver(filepath.Join(t.TempDir(), "missing", "audit.log"))
	assert.Error(t, err)
}
