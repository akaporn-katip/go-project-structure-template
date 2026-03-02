package customerprofile

import (
	"testing"
	"time"
)

func TestCreateCustomer(t *testing.T) {
	props := CreateCustomerProfileProps{
		Title:       "Mr.",
		Firstname:   "John",
		Lastname:    "Doe",
		DateOfBirth: "1990-01-01",
		Email:       "john.doe@example.com",
	}

	customer, err := CreateCustomer(props)

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}
	if customer == nil {
		t.Errorf("Expected customer but got nil")
		return
	}

	// Test getter methods
	if customer.Title() != "Mr." {
		t.Errorf("Expected title 'Mr.' but got %s", customer.Title())
	}
	if customer.Firstname() != "John" {
		t.Errorf("Expected firstname 'John' but got %s", customer.Firstname())
	}
	if customer.Lastname() != "Doe" {
		t.Errorf("Expected lastname 'Doe' but got %s", customer.Lastname())
	}
	if customer.Email().String() != "john.doe@example.com" {
		t.Errorf("Expected email 'john.doe@example.com' but got %s", customer.Email().String())
	}
	if customer.DateOfBirth().Year() != 1990 {
		t.Errorf("Expected birth year 1990 but got %d", customer.DateOfBirth().Year())
	}

	// Test business methods
	expectedFullName := "Mr. John Doe"
	if customer.FullName() != expectedFullName {
		t.Errorf("Expected full name '%s' but got %s", expectedFullName, customer.FullName())
	}

	expectedAge := time.Now().Year() - 1990
	age := customer.Age()
	if age != expectedAge && age != expectedAge-1 { // Account for birthday not yet passed
		t.Errorf("Expected age around %d but got %d", expectedAge, age)
	}

	if !customer.IsAdult() {
		t.Errorf("Expected customer to be adult")
	}

	// Test timestamps
	if customer.CreatedAt().IsZero() {
		t.Errorf("Expected createdAt to be set")
	}
	if customer.UpdatedAt().IsZero() {
		t.Errorf("Expected updatedAt to be set")
	}
	if customer.CreatedAt() != customer.UpdatedAt() {
		t.Errorf("Expected createdAt and updatedAt to be equal for new customer")
	}
}

func TestReconstructCustomer(t *testing.T) {
	id := "550e8400-e29b-41d4-a716-446655440000" // Sample UUID
	props := ReconstructCustomerProps{
		Title:       "Ms.",
		Firstname:   "Jane",
		Lastname:    "Smith",
		DateOfBirth: "1985-06-15",
		Email:       "jane.smith@example.com",
	}
	createdAt := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)

	customer, err := ReconstructCustomer(id, props, createdAt, updatedAt)

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}
	if customer == nil {
		t.Errorf("Expected customer but got nil")
		return
	}

	// Test getter methods
	if customer.ID().String() != id {
		t.Errorf("Expected ID '%s' but got %s", id, customer.ID().String())
	}
	if customer.Title() != "Ms." {
		t.Errorf("Expected title 'Ms.' but got %s", customer.Title())
	}
	if customer.Firstname() != "Jane" {
		t.Errorf("Expected firstname 'Jane' but got %s", customer.Firstname())
	}
	if customer.Lastname() != "Smith" {
		t.Errorf("Expected lastname 'Smith' but got %s", customer.Lastname())
	}

	// Test timestamps
	if !customer.CreatedAt().Equal(createdAt) {
		t.Errorf("Expected createdAt %v but got %v", createdAt, customer.CreatedAt())
	}
	if !customer.UpdatedAt().Equal(updatedAt) {
		t.Errorf("Expected updatedAt %v but got %v", updatedAt, customer.UpdatedAt())
	}

	// Test business methods
	expectedFullName := "Ms. Jane Smith"
	if customer.FullName() != expectedFullName {
		t.Errorf("Expected full name '%s' but got %s", expectedFullName, customer.FullName())
	}
}

func TestCustomerProfileBusinessMethods(t *testing.T) {
	// Test FullName without title
	props := CreateCustomerProfileProps{
		Title:       "", // Empty title
		Firstname:   "Bob",
		Lastname:    "Johnson",
		DateOfBirth: "1995-03-20",
		Email:       "bob.johnson@example.com",
	}

	customer, err := CreateCustomer(props)
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}

	expectedFullName := "Bob Johnson" // No title prefix
	if customer.FullName() != expectedFullName {
		t.Errorf("Expected full name '%s' but got %s", expectedFullName, customer.FullName())
	}

	// Test with minor (under 18)
	minorProps := CreateCustomerProfileProps{
		Title:       "Master",
		Firstname:   "Young",
		Lastname:    "Person",
		DateOfBirth: "2010-01-01", // 14 years old
		Email:       "young.person@example.com",
	}

	minor, err := CreateCustomer(minorProps)
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}

	if minor.IsAdult() {
		t.Errorf("Expected minor not to be adult")
	}

	age := minor.Age()
	expectedAge := time.Now().Year() - 2010
	if age != expectedAge && age != expectedAge-1 {
		t.Errorf("Expected age around %d but got %d", expectedAge, age)
	}
}
