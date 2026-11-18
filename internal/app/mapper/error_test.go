package mapper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorToGRPC(t *testing.T) {
	known := fmt.Errorf("wrapped: %w", entity.ErrNotFound)
	require.Equal(t, codes.NotFound, status.Code(ErrorToGRPC(known)))
	require.Equal(t, "not found", status.Convert(ErrorToGRPC(known)).Message())

	unknown := ErrorToGRPC(errors.New("db password leaked"))
	require.Equal(t, codes.Internal, status.Code(unknown))
	require.Equal(t, "internal error", status.Convert(unknown).Message())
}
