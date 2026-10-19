// Copyright IBM Corp. 2013, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	oslogin "google.golang.org/api/oslogin/v1"
)

// osLoginResponse is one canned reply from the fake OS Login API.
type osLoginResponse struct {
	status int
	body   string
}

const importSuccessBody = `{"loginProfile": {"name": "users/sa@example.com", "posixAccounts": [{"username": "sa_123", "primary": true}]}}`

func osLoginErrorBody(code int, message string) string {
	return fmt.Sprintf(`{"error": {"code": %d, "message": %q}}`, code, message)
}

// osLoginFake records the requests the fake OS Login API received.
type osLoginFake struct {
	mu       sync.Mutex
	requests []string // "METHOD path"
}

func (f *osLoginFake) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.requests)
}

// assertAllRequests checks every request went to the given method and path,
// so a test can't pass by hitting the wrong endpoint.
func (f *osLoginFake) assertAllRequests(t *testing.T, want string) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, got := range f.requests {
		assert.Equal(t, want, got)
	}
}

// newOSLoginTestDriver returns a driverGCE whose OS Login client talks to a
// local server that plays back responses in order, repeating the last one.
func newOSLoginTestDriver(t *testing.T, responses ...osLoginResponse) (*driverGCE, *osLoginFake) {
	t.Helper()
	fake := &osLoginFake{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.mu.Lock()
		fake.requests = append(fake.requests, r.Method+" "+r.URL.Path)
		n := len(fake.requests)
		fake.mu.Unlock()
		resp := responses[len(responses)-1]
		if n <= len(responses) {
			resp = responses[n-1]
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.status)
		_, _ = w.Write([]byte(resp.body))
	}))
	t.Cleanup(srv.Close)

	svc, err := oslogin.NewService(context.Background(),
		option.WithEndpoint(srv.URL), option.WithoutAuthentication())
	require.NoError(t, err)
	return &driverGCE{osLoginService: svc}, fake
}

const importRequest = "POST /v1/users/sa@example.com:importSshPublicKey"

func TestImportOSLoginSSHKey_NonRetryableErrorIsReturned(t *testing.T) {
	d, fake := newOSLoginTestDriver(t, osLoginResponse{
		status: http.StatusBadRequest,
		body:   osLoginErrorBody(400, "Login profile size exceeds 32 KiB. Delete profile values to make additional space."),
	})

	profile, err := d.ImportOSLoginSSHKey("sa@example.com", "ssh-ed25519 AAAA", nil)

	require.Error(t, err, "a failed import must not be reported as success")
	assert.Nil(t, profile)
	var apiErr *googleapi.Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.Code)
	assert.Equal(t, 1, fake.calls(), "non-409 errors must not be retried")
	fake.assertAllRequests(t, importRequest)
}

// noOSLoginRetrySleep makes the retry loop run without its 5-15s waits.
func noOSLoginRetrySleep(t *testing.T) {
	t.Helper()
	orig := osLoginRetrySleep
	osLoginRetrySleep = func(time.Duration) {}
	t.Cleanup(func() { osLoginRetrySleep = orig })
}

var conflict = osLoginResponse{
	status: http.StatusConflict,
	body:   osLoginErrorBody(409, "Multiple concurrent mutations were attempted. Please retry the request."),
}

func TestImportOSLoginSSHKey_Success(t *testing.T) {
	d, fake := newOSLoginTestDriver(t, osLoginResponse{status: http.StatusOK, body: importSuccessBody})

	profile, err := d.ImportOSLoginSSHKey("sa@example.com", "ssh-ed25519 AAAA", nil)

	require.NoError(t, err)
	require.NotNil(t, profile)
	require.Len(t, profile.PosixAccounts, 1)
	assert.Equal(t, "sa_123", profile.PosixAccounts[0].Username)
	assert.Equal(t, 1, fake.calls())
	fake.assertAllRequests(t, importRequest)
}

func TestImportOSLoginSSHKey_RetriesConflictThenSucceeds(t *testing.T) {
	noOSLoginRetrySleep(t)
	d, fake := newOSLoginTestDriver(t, conflict, osLoginResponse{status: http.StatusOK, body: importSuccessBody})

	profile, err := d.ImportOSLoginSSHKey("sa@example.com", "ssh-ed25519 AAAA", nil)

	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, 2, fake.calls())
	fake.assertAllRequests(t, importRequest)
}

func TestImportOSLoginSSHKey_PersistentConflictIsReturned(t *testing.T) {
	noOSLoginRetrySleep(t)
	d, fake := newOSLoginTestDriver(t, conflict)

	profile, err := d.ImportOSLoginSSHKey("sa@example.com", "ssh-ed25519 AAAA", nil)

	require.Error(t, err, "running out of retries must report the conflict, not success")
	assert.Nil(t, profile)
	var apiErr *googleapi.Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusConflict, apiErr.Code)
	assert.Equal(t, 10, fake.calls())
	fake.assertAllRequests(t, importRequest)
}
