package customerprofile

import (
	"time"
)

type CustomerProfile struct {
	id             CustomerID
	title          string
	firstname      string
	lastname       string
	dateOfBirth    DateOfBirth
	currentAddress Address
	email          Email
	identityCard   IdentityCard
	passport       Passport
	drivingLicense DrivingLicense

	createdAt time.Time
	updatedAt time.Time
}

type ReconstructCustomerProps struct {
	Title       string
	Firstname   string
	Lastname    string
	DateOfBirth string
	Email       string
}

type CreateCustomerProfileProps struct {
	Title       string
	Firstname   string
	Lastname    string
	DateOfBirth string
	Email       string
}

func ReconstructCustomer(id string, props ReconstructCustomerProps, createdAt time.Time, updatedAt time.Time) (*CustomerProfile, error) {
	dof, err := NewDateOfBirth(props.DateOfBirth)
	if err != nil {
		return nil, err
	}

	email, err := NewEmail(props.Email)
	if err != nil {
		return nil, err
	}

	customerId, err := NewCustomerID(id)
	if err != nil {
		return nil, err
	}

	return &CustomerProfile{
		id:          *customerId,
		title:       props.Title,
		firstname:   props.Firstname,
		lastname:    props.Lastname,
		email:       *email,
		dateOfBirth: *dof,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func CreateCustomer(props CreateCustomerProfileProps) (*CustomerProfile, error) {
	dof, err := NewDateOfBirth(props.DateOfBirth)
	if err != nil {
		return nil, err
	}

	email, err := NewEmail(props.Email)
	if err != nil {
		return nil, err
	}

	customerId := GenerateCustomerID()
	now := time.Now()

	return &CustomerProfile{
		id:          customerId,
		title:       props.Title,
		firstname:   props.Firstname,
		lastname:    props.Lastname,
		email:       *email,
		dateOfBirth: *dof,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func (c CustomerProfile) ChangeName(title string, firstname string, lastname string) error {

	return nil
}

// Getter methods
func (c CustomerProfile) ID() CustomerID {
	return c.id
}

func (c CustomerProfile) Title() string {
	return c.title
}

func (c CustomerProfile) Firstname() string {
	return c.firstname
}

func (c CustomerProfile) Lastname() string {
	return c.lastname
}

func (c CustomerProfile) DateOfBirth() DateOfBirth {
	return c.dateOfBirth
}

func (c CustomerProfile) CurrentAddress() Address {
	return c.currentAddress
}

func (c CustomerProfile) Email() Email {
	return c.email
}

func (c CustomerProfile) IdentityCard() IdentityCard {
	return c.identityCard
}

func (c CustomerProfile) Passport() Passport {
	return c.passport
}

func (c CustomerProfile) DrivingLicense() DrivingLicense {
	return c.drivingLicense
}

func (c CustomerProfile) CreatedAt() time.Time {
	return c.createdAt
}

func (c CustomerProfile) UpdatedAt() time.Time {
	return c.updatedAt
}

// Business methods
func (c CustomerProfile) FullName() string {
	if c.title != "" {
		return c.title + " " + c.firstname + " " + c.lastname
	}
	return c.firstname + " " + c.lastname
}

func (c CustomerProfile) Age() int {
	now := time.Now()
	age := now.Year() - c.dateOfBirth.Year()

	// For partial dates, we can only calculate approximate age
	if c.dateOfBirth.Precision() == FullDate {
		if now.YearDay() < c.dateOfBirth.Date().YearDay() {
			age--
		}
	}

	return age
}

func (c CustomerProfile) IsAdult() bool {
	return c.Age() >= 18
}
