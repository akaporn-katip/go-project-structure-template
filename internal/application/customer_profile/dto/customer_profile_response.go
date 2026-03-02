package dto

import "github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"

type CustomerIDResponse struct {
	ID string `json:"id"`
}

func ToCustomerIDResponse(id *customerprofile.CustomerID) CustomerIDResponse {
	return CustomerIDResponse{
		ID: id.String(),
	}
}
