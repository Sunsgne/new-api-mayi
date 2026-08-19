package service

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/stretchr/testify/require"
)

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

func TestTaskRelayRetryStatusCodeVendorRateLimitCode200(t *testing.T) {
	t.Parallel()

	require.Equal(t, http.StatusTooManyRequests, TaskRelayRetryStatusCode(&dto.TaskError{
		Code:       "200",
		StatusCode: http.StatusOK,
	}))
	require.Equal(t, http.StatusTooManyRequests, TaskRelayRetryStatusCode(&dto.TaskError{
		Code:       "200",
		StatusCode: http.StatusTooManyRequests,
	}))
	require.Equal(t, http.StatusOK, TaskRelayRetryStatusCode(&dto.TaskError{
		Code:       "invalid_request",
		StatusCode: http.StatusOK,
	}))
}
