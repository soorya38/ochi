package backend

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestDownloadBinaryHandler(t *testing.T) {
	// Setup
	osType := "linux"
	arch := "amd64"
	binaryName := "sensor-" + osType + "-" + arch
	binDir := "bin"

	// Create temp directory for the test workspace
	tmpDir := t.TempDir()

	// Save current WD and change to tmpDir
	originalWD, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir(tmpDir)
	assert.NoError(t, err)
	defer os.Chdir(originalWD)

	// Create bin directory in the temp workspace
	err = os.MkdirAll(binDir, 0755)
	assert.NoError(t, err)

	// Create a dummy binary file with placeholder UUID
	placeHolderUUID := "00000000-0000-0000-0000-000000000000"
	prefix := "some-prefix-bytes-"
	suffix := "-some-suffix-bytes"
	content := []byte(prefix + placeHolderUUID + suffix)
	filePath := filepath.Join(binDir, binaryName)
	err = os.WriteFile(filePath, content, 0644)
	assert.NoError(t, err)

	// Create server instance (empty is fine as handler uses no state)
	cs := &server{}

	// Create Request and Recorder
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/download/"+osType+"/"+arch, nil)

	// Mock params usually provided by router
	params := httprouter.Params{
		httprouter.Param{Key: "os", Value: osType},
		httprouter.Param{Key: "arch", Value: arch},
	}

	// Call handler directly
	cs.downloadBinaryHandler(w, r, params)

	// Assertions
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Verify headers
	assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment; filename=\"sensor-"+osType+"-"+arch+"\"")

	// Verify UUID replacement
	// The length should remain the same
	assert.Equal(t, len(content), len(body))

	// The placeholder should NOT be present
	assert.False(t, bytes.Contains(body, []byte(placeHolderUUID)))

	// The prefix and suffix should be present
	assert.True(t, bytes.Contains(body, []byte(prefix)))
	assert.True(t, bytes.Contains(body, []byte(suffix)))

	// Extract and verify the new UUID
	// We know the prefix length, so the UUID should start after it
	start := len(prefix)
	// UUID length is 36
	uuidLen := 36
	if len(body) >= start+uuidLen {
		extractedUUID := string(body[start : start+uuidLen])
		_, err := uuid.Parse(extractedUUID)
		assert.NoError(t, err, "The injected string should be a valid UUID")
	} else {
		assert.Fail(t, "Response body is too short to contain the UUID")
	}
}
