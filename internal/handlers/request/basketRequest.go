package request

import "errors"

type BasketRequest struct {
	ProductID int `json:"product_id"`
}

func (r *BasketRequest) Validate() error {
	if r.ProductID <= 0 {
		return errors.New("product_id must be greater than 0")
	}
	return nil
}
