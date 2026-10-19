package request

import (
	"errors"
	"osush/internal/validation"
)

type OrderRequest struct {
	CustomerName    *string    `json:"customer_name"`
	PhoneNumber     string     `json:"phone_number"`
	City            string     `json:"city"`
	DeliveryAddress string     `json:"delivery_address"`
	Notes           *string    `json:"notes"`
	CartItems       []CartItem `json:"cart_items"`
}

type CartItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (r *OrderRequest) Validate() error {
	if r.CustomerName == nil {
		return errors.New("customer_name is required")
	}

	if r.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}

	phoneNumber, err := validation.ValidatePhone(r.PhoneNumber)
	if err != nil {
		return err
	}

	r.PhoneNumber = phoneNumber

	if r.City == "" {
		return errors.New("city is required")
	}

	if r.DeliveryAddress == "" {
		return errors.New("delivery_address is required")
	}

	if r.Notes != nil {
		err := validation.ValidateTextSize(*r.Notes, "notes")
		if err != nil {
			return err
		}

	}

	return nil
}

type OrderListFilter struct {
	Search    string `form:"search"`
	IsSend    *bool  `form:"is_send"`
	Limit     string `form:"limit"`
	Offset    string `form:"offset"`
	LimitInt  int
	OffsetInt int
}
