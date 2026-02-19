package product

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrProductNameRequired),
		errors.Is(err, domain.ErrCategoryRequired),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidDiscountPercent),
		errors.Is(err, domain.ErrInvalidDiscountPeriod):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrProductNotActive),
		errors.Is(err, domain.ErrProductAlreadyActive),
		errors.Is(err, domain.ErrProductAlreadyInactive),
		errors.Is(err, domain.ErrProductArchived),
		errors.Is(err, domain.ErrDiscountNotActive),
		errors.Is(err, domain.ErrNoActiveDiscount):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}
