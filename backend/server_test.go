package backend

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestDownloadBinaryHandler(t *testing.T) {
	// Setup
	osType := "linux"
	arch := "amd64"
	binaryName := "sensor-" + osType + "-" + arch
	binDir := "bin"

	// Create bin directory
	err := os.MkdirAll(binDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(binDir) // Cleanup

	// Create a dummy binary file with placeholder UUID
	placeHolderUUID := "00000000-0000-0000-0000-000000000000"
	content := []byte("some-prefix-bytes-" + placeHolderUUID + "-some-suffix-bytes")
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
	assert.True(t, bytes.Contains(body, []byte("some-prefix-bytes-")))
	assert.True(t, bytes.Contains(body, []byte("-some-suffix-bytes")))
}
