package grpcutils

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewBadRequestError(mapErr map[string]string) error {
	st := status.New(codes.InvalidArgument, "Request failed validation")

	errDetails := &errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{},
	}

	for field, errMsg := range mapErr {
		errDetails.FieldViolations = append(errDetails.FieldViolations,
			&errdetails.BadRequest_FieldViolation{
				Field:       field,
				Description: errMsg,
			})
	}

	stWithDetails, err := st.WithDetails(errDetails)
	if err != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}
