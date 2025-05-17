// Package golden contains test helpers for reading data from ./testdata/ subdirectory.
package golden

import (
	"bytes"
	"io"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Open file and close on test cleanup.
func Open(t *testing.T, name string) io.ReadSeeker {
	t.Helper()

	f, err := os.Open(path.Join("testdata", name))
	require.NoError(t, err)

	t.Cleanup(func() { f.Close() })

	return f
}

// ReadString reads file into string.
func ReadString(t *testing.T, name string) string {
	t.Helper()

	buf := strings.Builder{}
	_, err := io.Copy(&buf, Open(t, name))
	require.NoError(t, err)

	return buf.String()
}

// ReadBytes reads file into []byte.
func ReadBytes(t *testing.T, name string) []byte {
	t.Helper()

	buf := bytes.Buffer{}
	_, err := io.Copy(&buf, Open(t, name))
	require.NoError(t, err)

	return buf.Bytes()
}
