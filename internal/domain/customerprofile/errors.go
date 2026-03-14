package customerprofile

import "github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"

var (
	ErrInvalidDateOfBirthFormat = domainerrors.NewInvalidInputError("date of birth must be in format YYYY-MM-DD, YYYY-MM, or YYYY")
	ErrFutureBirthDate          = domainerrors.NewInvalidInputError("birth date cannot be in the future")
	ErrTooOldBirthDate          = domainerrors.NewInvalidInputError("birth date is unreasonably old")

	ErrInvalidThaiIDNumber  = domainerrors.NewInvalidInputError("thai ID number must be 13 digits with valid checksum")
	ErrInvalidFirstname     = domainerrors.NewInvalidInputError("first name cannot be empty")
	ErrInvalidLastname      = domainerrors.NewInvalidInputError("last name cannot be empty")
	ErrInvalidIssueDate     = domainerrors.NewInvalidInputError("issue date must be in format YYYY-MM-DD")
	ErrInvalidExpiryDate    = domainerrors.NewInvalidInputError("expiry date must be in format YYYY-MM-DD")
	ErrIssueDateAfterExpiry = domainerrors.NewInvalidInputError("issue date cannot be after expiry date")

	ErrInvalidPassportNumber        = domainerrors.NewInvalidInputError("passport number cannot be empty")
	ErrInvalidIssuingCountry        = domainerrors.NewInvalidInputError("issuing country cannot be empty")
	ErrInvalidPassportIssueDate     = domainerrors.NewInvalidInputError("passport issue date must be in format YYYY-MM-DD")
	ErrInvalidPassportExpiryDate    = domainerrors.NewInvalidInputError("passport expiry date must be in format YYYY-MM-DD")
	ErrPassportIssueDateAfterExpiry = domainerrors.NewInvalidInputError("passport issue date cannot be after expiry date")
	ErrPassportFutureIssueDate      = domainerrors.NewInvalidInputError("passport issue date cannot be in the future")
	ErrPassportTooOld               = domainerrors.NewInvalidInputError("passport is too old to be valid")

	ErrInvalidDrivingLicenseNumber        = domainerrors.NewInvalidInputError("driving license number cannot be empty")
	ErrInvalidDrivingLicenseClass         = domainerrors.NewInvalidInputError("invalid driving license class")
	ErrInvalidIssuingAuthority            = domainerrors.NewInvalidInputError("issuing authority cannot be empty")
	ErrInvalidDrivingLicenseIssueDate     = domainerrors.NewInvalidInputError("driving license issue date must be in format YYYY-MM-DD")
	ErrInvalidDrivingLicenseExpiryDate    = domainerrors.NewInvalidInputError("driving license expiry date must be in format YYYY-MM-DD")
	ErrDrivingLicenseIssueDateAfterExpiry = domainerrors.NewInvalidInputError("driving license issue date cannot be after expiry date")
	ErrDrivingLicenseFutureIssueDate      = domainerrors.NewInvalidInputError("driving license issue date cannot be in the future")

	ErrInvalidEmailFormat = domainerrors.NewInvalidInputError("invalid email format")

	ErrInvalidAddressHouseNumber = domainerrors.NewInvalidInputError("house number is required")
	ErrInvalidAddressSubdistrict = domainerrors.NewInvalidInputError("subdistrict is required")
	ErrInvalidAddressDistrict    = domainerrors.NewInvalidInputError("district is required")
	ErrInvalidAddressProvince    = domainerrors.NewInvalidInputError("province is required")
	ErrInvalidAddressPostalCode  = domainerrors.NewInvalidInputError("postal code is required")
	ErrInvalidThaiPostalCode     = domainerrors.NewInvalidInputError("thai postal code must be 5 digits")

	ErrInvalidUUIDFormat = domainerrors.NewInvalidInputError("invalid UUID format")

	ErrInvalidPhoneNumberFormat = domainerrors.NewInvalidInputError("invalid phonenumber format")
)

func NewEmailAlreadyExistsError(email string) *domainerrors.DomainError {
	return domainerrors.NewAlreadyExistsError("email already exists").
		WithDetail("email", email)
}

func NewFindByEmailNotFoundError(email string) *domainerrors.DomainError {
	return domainerrors.NewNotFoundError("customer profile by email not found").WithDetail("email", email)
}

func NewFindByIDNotFoundError() *domainerrors.DomainError {
	return domainerrors.NewNotFoundError("customer profile by id not found")
}
