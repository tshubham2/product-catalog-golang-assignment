package domain

import "errors"

var (
	ErrProductNotFound        = errors.New("product not found")
	ErrProductNotActive       = errors.New("product is not active")
	ErrProductAlreadyActive   = errors.New("product is already active")
	ErrProductAlreadyInactive = errors.New("product is already inactive")
	ErrProductArchived        = errors.New("product is archived")
	ErrInvalidPrice           = errors.New("price must be positive")
	ErrInvalidDiscountPercent = errors.New("discount percentage must be between 0 and 100 exclusive")
	ErrInvalidDiscountPeriod  = errors.New("discount start date must be before end date")
	ErrDiscountNotActive      = errors.New("discount is not active at the given time")
	ErrNoActiveDiscount       = errors.New("product has no discount to remove")
	ErrProductNameRequired    = errors.New("product name is required")
	ErrCategoryRequired       = errors.New("product category is required")
)
