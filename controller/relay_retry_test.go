package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestShouldRetryWrappedVendorRateLimitCode200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name       string
		statusCode int
		errorCode  any
		want       bool
	}{
		{
			name:       "retry http 429 with wrapped code 200",
			statusCode: http.StatusTooManyRequests,
			errorCode:  200,
			want:       true,
		},
		{
			name:       "retry http 200 with vendor code 200",
			statusCode: http.StatusOK,
			errorCode:  200,
			want:       true,
		},
		{
			name:       "no retry http 200 with unrelated error",
			statusCode: http.StatusOK,
			errorCode:  "invalid_request",
			want:       false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			apiErr := types.WithOpenAIError(types.OpenAIError{
				Message: "rate limit exceeded",
				Type:    "upstream_error",
				Code:    tc.errorCode,
			}, tc.statusCode)

			require.Equal(t, tc.want, shouldRetry(c, apiErr, 1))
		})
	}
}
