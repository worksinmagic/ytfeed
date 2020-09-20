package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Run("Health check success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		res := httptest.NewRecorder()

		Handler(res, req)

		require.Equal(t, http.StatusOK, res.Result().StatusCode)
	})
}
