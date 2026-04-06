//go:build integration && !no_e2e

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

func TestDebugStorageUpload2(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureStorageSchema()
	tc.Config.Storage.Provider = "local"
	tc.Config.Storage.LocalPath = "/tmp/fluxbase-test-storage-debug2"
	tc.CleanupStorageFiles()
	defer tc.Close()

	serviceKey := tc.CreateServiceKey("Debug2 Service Key")
	apiKey := tc.CreateAPIKey("Debug2 API Key", nil)

	bucketName := "upload-debug-test-12345"
	fileName := "test.txt"
	fileContent := []byte("Hello, World!")

	// Create bucket
	tc.NewRequest("POST", "/api/v1/storage/buckets/"+bucketName).
		WithServiceKey(serviceKey).
		Send().
		AssertStatus(fiber.StatusCreated)

	// Upload with API key
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	require.NoError(t, err)
	_, err = part.Write(fileContent)
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/storage/"+bucketName+"/"+fileName, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Client-Key", apiKey)

	resp, err := tc.App.Test(req)
	require.NoError(t, err)
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("API key - Status: %d, Body: %s\n", resp.StatusCode, string(respBody))

	// Also try with service key
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	part2, err := writer2.CreateFormFile("file", "test2.txt")
	require.NoError(t, err)
	_, err = part2.Write(fileContent)
	require.NoError(t, err)
	err = writer2.Close()
	require.NoError(t, err)

	req2 := httptest.NewRequest("POST", "/api/v1/storage/"+bucketName+"/test2.txt", body2)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	req2.Header.Set("X-Service-Key", serviceKey)

	resp2, err := tc.App.Test(req2)
	require.NoError(t, err)
	respBody2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Service key - Status: %d, Body: %s\n", resp2.StatusCode, string(respBody2))

	// Try with no auth at all
	body3 := &bytes.Buffer{}
	writer3 := multipart.NewWriter(body3)
	part3, err := writer3.CreateFormFile("file", "test3.txt")
	require.NoError(t, err)
	_, err = part3.Write(fileContent)
	require.NoError(t, err)
	err = writer3.Close()
	require.NoError(t, err)

	req3 := httptest.NewRequest("POST", "/api/v1/storage/"+bucketName+"/test3.txt", body3)
	req3.Header.Set("Content-Type", writer3.FormDataContentType())

	resp3, err := tc.App.Test(req3)
	require.NoError(t, err)
	respBody3, _ := io.ReadAll(resp3.Body)
	fmt.Printf("No auth - Status: %d, Body: %s\n", resp3.StatusCode, string(respBody3))
}
