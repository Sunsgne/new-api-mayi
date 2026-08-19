package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	taskdto "github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/require"
)

// TestRelayErrorHandlerWrappedVendorRateLimitCode200 exercises the real
// upstream-JSON parsing path (RelayErrorHandler -> RelayRetryStatusCode),
// not just a hand-built NewAPIError, since production traffic always goes
// through JSON unmarshaling where numeric error codes decode as float64.
func TestRelayErrorHandlerWrappedVendorRateLimitCode200(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		httpStatus     int
		body           string
		wantRetryCode  int
		wantHTTPStatus int
	}{
		{
			name:           "HTTP 429 with numeric wrapped code 200",
			httpStatus:     http.StatusTooManyRequests,
			body:           `{"error":{"message":"rate limited","type":"upstream_error","code":200}}`,
			wantRetryCode:  http.StatusTooManyRequests,
			wantHTTPStatus: http.StatusTooManyRequests,
		},
		{
			name:           "HTTP 429 with string wrapped code 200",
			httpStatus:     http.StatusTooManyRequests,
			body:           `{"error":{"message":"rate limited","type":"upstream_error","code":"200"}}`,
			wantRetryCode:  http.StatusTooManyRequests,
			wantHTTPStatus: http.StatusTooManyRequests,
		},
		{
			name:           "HTTP 429 with unrelated code stays 429",
			httpStatus:     http.StatusTooManyRequests,
			body:           `{"error":{"message":"rate limited","type":"upstream_error","code":"rate_limit_exceeded"}}`,
			wantRetryCode:  http.StatusTooManyRequests,
			wantHTTPStatus: http.StatusTooManyRequests,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &http.Response{
				StatusCode: tc.httpStatus,
				Body:       io.NopCloser(strings.NewReader(tc.body)),
			}

			apiErr := RelayErrorHandler(context.Background(), resp, false)

			require.NotNil(t, apiErr)
			require.Equal(t, tc.wantHTTPStatus, apiErr.StatusCode, "client-facing status must remain unchanged")
			require.Equal(t, tc.wantRetryCode, RelayRetryStatusCode(apiErr))
		})
	}
}

func TestRelayRetryStatusCodeVendorRateLimitCode200(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		statusCode int
		errorCode  any
		want       int
	}{
		{
			name:       "http 429 with wrapped vendor code 200",
			statusCode: http.StatusTooManyRequests,
			errorCode:  200,
			want:       http.StatusTooManyRequests,
		},
		{
			name:       "http 200 with vendor code 200",
			statusCode: http.StatusOK,
			errorCode:  200,
			want:       http.StatusTooManyRequests,
		},
		{
			name:       "http 200 with vendor code 200 string",
			statusCode: http.StatusOK,
			errorCode:  "200",
			want:       http.StatusTooManyRequests,
		},
		{
			name:       "http 200 with unrelated error code",
			statusCode: http.StatusOK,
			errorCode:  "content_moderation_failed",
			want:       http.StatusOK,
		},
		{
			name:       "http 502 without vendor code",
			statusCode: http.StatusBadGateway,
			errorCode:  "server_error",
			want:       http.StatusBadGateway,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			apiErr := types.WithOpenAIError(types.OpenAIError{
				Message: "rate limit exceeded",
				Type:    "upstream_error",
				Code:    tc.errorCode,
			}, tc.statusCode)

			require.Equal(t, tc.want, RelayRetryStatusCode(apiErr))
		})
	}
}

// TestHTTP200BodyWrappedVendorRateLimitCode200 covers the other real-world
// shape from the task: some upstreams answer with HTTP 200 and embed the
// rate-limit signal purely in the response body's error.code. This mirrors
// how OpenAI-compatible adaptors parse body errors via dto.GetOpenAIError.
func TestHTTP200BodyWrappedVendorRateLimitCode200(t *testing.T) {
	t.Parallel()

	var simpleResponse dto.SimpleResponse
	body := `{"error":{"message":"rate limited","type":"upstream_error","code":200}}`
	require.NoError(t, common.Unmarshal([]byte(body), &simpleResponse))

	oaiError := simpleResponse.GetOpenAIError()
	require.NotNil(t, oaiError)

	apiErr := types.WithOpenAIError(*oaiError, http.StatusOK)

	require.Equal(t, http.StatusOK, apiErr.StatusCode, "client-facing status must remain 200")
	require.Equal(t, http.StatusTooManyRequests, RelayRetryStatusCode(apiErr))
}

func TestTaskRelayRetryStatusCodeVendorRateLimitCode200(t *testing.T) {
	t.Parallel()

	require.Equal(t, http.StatusTooManyRequests, TaskRelayRetryStatusCode(&taskdto.TaskError{
		Code:       "200",
		StatusCode: http.StatusOK,
	}))
	require.Equal(t, http.StatusTooManyRequests, TaskRelayRetryStatusCode(&taskdto.TaskError{
		Code:       "200",
		StatusCode: http.StatusTooManyRequests,
	}))
	require.Equal(t, http.StatusOK, TaskRelayRetryStatusCode(&taskdto.TaskError{
		Code:       "invalid_request",
		StatusCode: http.StatusOK,
	}))
}
