package entity

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestAppErrorTransportCodes(t *testing.T) {
	cases := []struct {
		kind       ErrorKind
		httpStatus int
		grpcCode   codes.Code
	}{
		{KindNotFound, http.StatusNotFound, codes.NotFound},
		{KindAlreadyExists, http.StatusConflict, codes.AlreadyExists},
		{KindFailedPrecondition, http.StatusConflict, codes.FailedPrecondition},
		{KindInvalidArgument, http.StatusBadRequest, codes.InvalidArgument},
		{KindUnknown, http.StatusInternalServerError, codes.Internal},
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			err := NewAppError(tc.kind, "message")
			require.Equal(t, tc.httpStatus, err.HTTPStatus())
			require.Equal(t, tc.grpcCode, err.GRPCCode())
		})
	}
}
