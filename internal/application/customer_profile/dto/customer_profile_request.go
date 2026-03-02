package dto

type CreateCustomerProfileRequest struct {
	Title       string `json:"title" validate:"required"`
	Firstname   string `json:"firstname" validate:"required"`
	Lastname    string `json:"lastname" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	DateOfBirth string `json:"dateOfBirth" validate:"required"`
}
