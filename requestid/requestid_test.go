package requestid_test

import (
	"testing"

	"github.com/DKhorkov/libs/requestid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should return a new request ID", func(t *testing.T) {
		requestID := requestid.New()
		assert.NotEmpty(t, requestID)
		err := uuid.Validate(requestID)
		require.NoError(t, err)
	})
}
