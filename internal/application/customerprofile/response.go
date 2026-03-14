package customerprofileapp

import "github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"

type CustomerIDResponse struct {
	ID string `json:"id"`
}

func ToCustomerIDResponse(id *customerprofile.CustomerID) CustomerIDResponse {
	return CustomerIDResponse{
		ID: id.String(),
	}
}

type CustomerResponse struct {
	CustomerIDResponse
	Title       string `json:"title"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	DateOfBirth string `json:"dateOfBirth"`
	// currentAddress Address
	Email          string `json:"email"`
	IdentityCard   string `json:"identityCard"`
	Passport       string `json:"passport"`
	DrivingLicense string `json:"drivingLicense"`
}

func ToCustomerResponse(customer *customerprofile.CustomerProfile) CustomerResponse {
	id := customer.ID()
	return CustomerResponse{
		CustomerIDResponse: ToCustomerIDResponse(&id),
		Email:              customer.Email().String(),
		IdentityCard:       customer.IdentityCard().ID(),
		Firstname:          customer.Firstname(),
		Lastname:           customer.Lastname(),
		Title:              customer.Title(),
		DateOfBirth:        customer.DateOfBirth().ISOString(),
	}
}
