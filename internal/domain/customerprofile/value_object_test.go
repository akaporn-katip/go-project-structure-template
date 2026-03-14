package customerprofile

import (
	"strings"
	"testing"
	"time"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			expectError: true,
		},
		{
			name:        "invalid email without @",
			email:       "testexample.com",
			expectError: true,
		},
		{
			name:        "invalid email without domain",
			email:       "test@",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if email != nil {
					t.Errorf("Expected nil email but got %v", email)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
					return
				}
				if email == nil {
					t.Errorf("Expected email but got nil")
					return
				}
				if email.String() != strings.ToLower(tt.email) {
					t.Errorf("Expected %s but got %s", strings.ToLower(tt.email), email.String())
				}
			}
		})
	}
}

func TestNewDateOfBirth(t *testing.T) {
	tests := []struct {
		name        string
		dob         string
		expectError bool
		precision   DatePrecision
	}{
		{
			name:        "full date",
			dob:         "19901225",
			expectError: false,
			precision:   FullDate,
		},
		{
			name:        "year and month",
			dob:         "199012",
			expectError: false,
			precision:   MonthYear,
		},
		{
			name:        "year only",
			dob:         "1990",
			expectError: false,
			precision:   YearOnly,
		},
		{
			name:        "future date",
			dob:         "20300101",
			expectError: true,
		},
		{
			name:        "too old date",
			dob:         "18000101",
			expectError: true,
		},
		{
			name:        "invalid format",
			dob:         "invalid-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob, err := NewDateOfBirth(tt.dob)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if dob != nil {
					t.Errorf("Expected nil dob but got %v", dob)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
					return
				}
				if dob == nil {
					t.Errorf("Expected dob but got nil")
					return
				}
				if dob.Precision() != tt.precision {
					t.Errorf("Expected precision %v but got %v", tt.precision, dob.Precision())
				}
			}
		})
	}
}

func TestDateOfBirthMethods(t *testing.T) {
	dob, err := NewDateOfBirth("19901225")
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

	if dob.Year() != 1990 {
		t.Errorf("Expected year 1990 but got %d", dob.Year())
	}

	if dob.Month() != time.December {
		t.Errorf("Expected month December but got %v", dob.Month())
	}

	if dob.Day() != 25 {
		t.Errorf("Expected day 25 but got %d", dob.Day())
	}

	expectedString := "19901225"
	if dob.String() != expectedString {
		t.Errorf("Expected string %s but got %s", expectedString, dob.String())
	}
}

func TestAddressBuilder(t *testing.T) {
	address, err := NewAddressBuilder().
		WithHouseNumber("456").
		WithBuilding("Test Building").
		WithSoi("Sukhumvit 21").
		WithRoad("Sukhumvit").
		WithSubdistrict("Khlong Toei Nuea").
		WithDistrict("Watthana").
		WithProvince("Bangkok").
		WithPostalCode("10110").
		WithCountry("Thailand").
		Build()

	if err != nil {
		t.Errorf("Expected no error but got %v", err)
		return
	}
	if address == nil {
		t.Errorf("Expected address but got nil")
		return
	}
	if address.HouseNumber() != "456" {
		t.Errorf("Expected house number 456 but got %s", address.HouseNumber())
	}
	if address.Building() != "Test Building" {
		t.Errorf("Expected building 'Test Building' but got %s", address.Building())
	}
	if !strings.Contains(address.FullAddress(), "ซอย Sukhumvit 21") {
		t.Errorf("Expected full address to contain 'ซอย Sukhumvit 21' but got %s", address.FullAddress())
	}
}

func TestAddressBuilderValidation(t *testing.T) {
	// Test missing required fields
	address, err := NewAddressBuilder().
		WithBuilding("Test Building").
		Build()

	if err == nil {
		t.Errorf("Expected error for missing house number")
	}
	if address != nil {
		t.Errorf("Expected nil address but got %v", address)
	}
}

func TestIsValidThaiPostalCode(t *testing.T) {
	tests := []struct {
		code  string
		valid bool
	}{
		{"10110", true},
		{"12345", true},
		{"1234", false},   // too short
		{"123456", false}, // too long
		{"1234a", false},  // contains letter
		{"", false},       // empty
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := isValidThaiPostalCode(tt.code)
			if result != tt.valid {
				t.Errorf("Expected %v for postal code %s but got %v", tt.valid, tt.code, result)
			}
		})
	}
}

func TestNewIdentityCard(t *testing.T) {
	validAddress, _ := NewAddressBuilder().
		WithHouseNumber("123").
		WithSubdistrict("Khlong Toei").
		WithDistrict("Khlong Toei").
		WithProvince("Bangkok").
		WithPostalCode("10110").
		Build()

	validInput := IdentityCardInput{
		Number:      "1320500269543", // Valid Thai ID with correct checksum
		Title:       "Mr.",
		Firstname:   "John",
		Lastname:    "Doe",
		DateOfBirth: "19900101",
		IssueDate:   "2020-01-01",
		ExpiryDate:  "2030-01-01",
		Address:     *validAddress,
	}

	t.Run("valid identity card", func(t *testing.T) {
		idCard, err := NewIdentityCard(validInput)

		if err != nil {
			t.Errorf("Expected no error but got %v", err)
			return
		}
		if idCard == nil {
			t.Errorf("Expected identity card but got nil")
			return
		}
		if idCard.number != "1320500269543" {
			t.Errorf("Expected number 1320500269543 but got %s", idCard.number)
		}
		if idCard.firstname != "John" {
			t.Errorf("Expected firstname John but got %s", idCard.firstname)
		}
		if idCard.lastname != "Doe" {
			t.Errorf("Expected lastname Doe but got %s", idCard.lastname)
		}
		if idCard.title != "Mr." {
			t.Errorf("Expected title Mr. but got %s", idCard.title)
		}
	})

	t.Run("empty firstname", func(t *testing.T) {
		input := validInput
		input.Firstname = ""

		idCard, err := NewIdentityCard(input)

		if err == nil {
			t.Errorf("Expected error for empty firstname")
		}
		if idCard != nil {
			t.Errorf("Expected nil identity card but got %v", idCard)
		}
	})

	t.Run("empty lastname", func(t *testing.T) {
		input := validInput
		input.Lastname = ""

		idCard, err := NewIdentityCard(input)

		if err == nil {
			t.Errorf("Expected error for empty lastname")
		}
		if idCard != nil {
			t.Errorf("Expected nil identity card but got %v", idCard)
		}
	})

	t.Run("invalid ID number", func(t *testing.T) {
		input := validInput
		input.Number = "123456789012" // 12 digits instead of 13

		idCard, err := NewIdentityCard(input)

		if err == nil {
			t.Errorf("Expected error for invalid ID number")
		}
		if idCard != nil {
			t.Errorf("Expected nil identity card but got %v", idCard)
		}
	})

	t.Run("issue date after expiry", func(t *testing.T) {
		input := validInput
		input.IssueDate = "2030-01-01"
		input.ExpiryDate = "2025-01-01"

		idCard, err := NewIdentityCard(input)

		if err == nil {
			t.Errorf("Expected error for issue date after expiry")
		}
		if idCard != nil {
			t.Errorf("Expected nil identity card but got %v", idCard)
		}
	})

	t.Run("invalid date format", func(t *testing.T) {
		input := validInput
		input.IssueDate = "invalid-date"

		idCard, err := NewIdentityCard(input)

		if err == nil {
			t.Errorf("Expected error for invalid date format")
		}
		if idCard != nil {
			t.Errorf("Expected nil identity card but got %v", idCard)
		}
	})
}

func TestNewPassport(t *testing.T) {
	validInput := PassportInput{
		Number:         "AB1234567",
		IssuedDate:     "2020-01-01",
		ExpiryDate:     "2030-01-01",
		IssuingCountry: "Thailand",
	}

	t.Run("valid passport", func(t *testing.T) {
		passport, err := NewPassport(validInput)

		if err != nil {
			t.Errorf("Expected no error but got %v", err)
			return
		}
		if passport == nil {
			t.Errorf("Expected passport but got nil")
			return
		}
		if passport.Number() != "AB1234567" {
			t.Errorf("Expected number AB1234567 but got %s", passport.Number())
		}
		if passport.IssuingCountry() != "Thailand" {
			t.Errorf("Expected issuing country Thailand but got %s", passport.IssuingCountry())
		}
		if passport.IsExpired() {
			t.Errorf("Expected passport not to be expired")
		}
		if !passport.IsValid() {
			t.Errorf("Expected passport to be valid")
		}
	})

	t.Run("empty passport number", func(t *testing.T) {
		input := validInput
		input.Number = ""

		passport, err := NewPassport(input)

		if err == nil {
			t.Errorf("Expected error for empty passport number")
		}
		if passport != nil {
			t.Errorf("Expected nil passport but got %v", passport)
		}
	})

	t.Run("empty issuing country", func(t *testing.T) {
		input := validInput
		input.IssuingCountry = ""

		passport, err := NewPassport(input)

		if err == nil {
			t.Errorf("Expected error for empty issuing country")
		}
		if passport != nil {
			t.Errorf("Expected nil passport but got %v", passport)
		}
	})

	t.Run("future issue date", func(t *testing.T) {
		input := validInput
		futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		input.IssuedDate = futureDate

		passport, err := NewPassport(input)

		if err == nil {
			t.Errorf("Expected error for future issue date")
		}
		if passport != nil {
			t.Errorf("Expected nil passport but got %v", passport)
		}
	})

	t.Run("issue date after expiry", func(t *testing.T) {
		input := validInput
		input.IssuedDate = "2030-01-01"
		input.ExpiryDate = "2025-01-01"

		passport, err := NewPassport(input)

		if err == nil {
			t.Errorf("Expected error for issue date after expiry")
		}
		if passport != nil {
			t.Errorf("Expected nil passport but got %v", passport)
		}
	})
}

func TestPassportBusinessMethods(t *testing.T) {
	// Create expired passport
	expiredPassport, _ := NewPassport(PassportInput{
		Number:         "EX123456",
		IssuedDate:     "2010-01-01",
		ExpiryDate:     "2020-01-01", // Expired
		IssuingCountry: "Thailand",
	})

	if !expiredPassport.IsExpired() {
		t.Errorf("Expected passport to be expired")
	}
	if expiredPassport.IsValid() {
		t.Errorf("Expected expired passport to be invalid")
	}
	if expiredPassport.DaysUntilExpiry() != 0 {
		t.Errorf("Expected 0 days until expiry for expired passport")
	}

	// Create valid passport
	validPassport, _ := NewPassport(PassportInput{
		Number:         "VA123456",
		IssuedDate:     "2020-01-01",
		ExpiryDate:     "2030-01-01",
		IssuingCountry: "Thailand",
	})

	if !validPassport.IsThaiPassport() {
		t.Errorf("Expected Thai passport to be identified as Thai")
	}

	// Test string representation
	expectedString := "Passport VA123456 (Thailand) - Valid until 2030-01-01"
	if validPassport.String() != expectedString {
		t.Errorf("Expected string %s but got %s", expectedString, validPassport.String())
	}
}

func TestNewDrivingLicense(t *testing.T) {
	validInput := DrivingLicenseInput{
		Number:           "DL123456789",
		IssuedDate:       "2020-01-01",
		ExpiryDate:       "2025-01-01",
		Class:            "B1",
		IssuingAuthority: "Department of Land Transport",
	}

	t.Run("valid driving license", func(t *testing.T) {
		dl, err := NewDrivingLicense(validInput)

		if err != nil {
			t.Errorf("Expected no error but got %v", err)
			return
		}
		if dl == nil {
			t.Errorf("Expected driving license but got nil")
			return
		}
		if dl.Number() != "DL123456789" {
			t.Errorf("Expected number DL123456789 but got %s", dl.Number())
		}
		if dl.Class() != "B1" {
			t.Errorf("Expected class B1 but got %s", dl.Class())
		}
		if !dl.CanDriveCar() {
			t.Errorf("Expected B1 license to allow driving car")
		}
		if dl.CanDriveMotorcycle() {
			t.Errorf("Expected B1 license not to allow driving motorcycle")
		}
		if dl.GetVehicleType() != "Car" {
			t.Errorf("Expected vehicle type Car but got %s", dl.GetVehicleType())
		}
	})

	t.Run("motorcycle license", func(t *testing.T) {
		input := validInput
		input.Class = "A1"

		dl, err := NewDrivingLicense(input)

		if err != nil {
			t.Errorf("Expected no error but got %v", err)
			return
		}
		if !dl.CanDriveMotorcycle() {
			t.Errorf("Expected A1 license to allow driving motorcycle")
		}
		if dl.CanDriveCar() {
			t.Errorf("Expected A1 license not to allow driving car")
		}
		if dl.GetVehicleType() != "Motorcycle" {
			t.Errorf("Expected vehicle type Motorcycle but got %s", dl.GetVehicleType())
		}
	})

	t.Run("truck license", func(t *testing.T) {
		input := validInput
		input.Class = "C1"

		dl, err := NewDrivingLicense(input)

		if err != nil {
			t.Errorf("Expected no error but got %v", err)
			return
		}
		if !dl.CanDriveTruck() {
			t.Errorf("Expected C1 license to allow driving truck")
		}
		if dl.GetVehicleType() != "Truck" {
			t.Errorf("Expected vehicle type Truck but got %s", dl.GetVehicleType())
		}
	})

	t.Run("invalid license class", func(t *testing.T) {
		input := validInput
		input.Class = "INVALID"

		dl, err := NewDrivingLicense(input)

		if err == nil {
			t.Errorf("Expected error for invalid license class")
		}
		if dl != nil {
			t.Errorf("Expected nil driving license but got %v", dl)
		}
	})

	t.Run("empty license number", func(t *testing.T) {
		input := validInput
		input.Number = ""

		dl, err := NewDrivingLicense(input)

		if err == nil {
			t.Errorf("Expected error for empty license number")
		}
		if dl != nil {
			t.Errorf("Expected nil driving license but got %v", dl)
		}
	})

	t.Run("empty issuing authority", func(t *testing.T) {
		input := validInput
		input.IssuingAuthority = ""

		dl, err := NewDrivingLicense(input)

		if err == nil {
			t.Errorf("Expected error for empty issuing authority")
		}
		if dl != nil {
			t.Errorf("Expected nil driving license but got %v", dl)
		}
	})

	t.Run("future issue date", func(t *testing.T) {
		input := validInput
		futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
		input.IssuedDate = futureDate

		dl, err := NewDrivingLicense(input)

		if err == nil {
			t.Errorf("Expected error for future issue date")
		}
		if dl != nil {
			t.Errorf("Expected nil driving license but got %v", dl)
		}
	})
}

func TestDrivingLicenseBusinessMethods(t *testing.T) {
	// Create expired license
	expiredLicense, _ := NewDrivingLicense(DrivingLicenseInput{
		Number:           "EX123456",
		IssuedDate:       "2010-01-01",
		ExpiryDate:       "2020-01-01", // Expired
		Class:            "B1",
		IssuingAuthority: "Department of Land Transport",
	})

	if !expiredLicense.IsExpired() {
		t.Errorf("Expected license to be expired")
	}
	if expiredLicense.IsValid() {
		t.Errorf("Expected expired license to be invalid")
	}
	if expiredLicense.DaysUntilExpiry() != 0 {
		t.Errorf("Expected 0 days until expiry for expired license")
	}
	if !expiredLicense.IsExpiringSoon(30) {
		t.Errorf("Expected expired license to be expiring soon")
	}

	// Test string representation
	expectedString := "Driving License EX123456 (Class B1) - Valid until 2020-01-01"
	if expiredLicense.String() != expectedString {
		t.Errorf("Expected string %s but got %s", expectedString, expiredLicense.String())
	}
}

func TestIsValidThaiLicenseClass(t *testing.T) {
	tests := []struct {
		class string
		valid bool
	}{
		{"A1", true},
		{"A2", true},
		{"A3", true},
		{"A", true},
		{"B1", true},
		{"B2", true},
		{"B3", true},
		{"B", true},
		{"C1", true},
		{"C2", true},
		{"C3", true},
		{"C", true},
		{"D1", true},
		{"D2", true},
		{"D3", true},
		{"D", true},
		{"E1", true},
		{"E2", true},
		{"E3", true},
		{"E", true},
		{"a1", true}, // should be case insensitive
		{"b1", true}, // should be case insensitive
		{"INVALID", false},
		{"Z1", false},
		{"", false},
		{"F1", false},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			result := isValidThaiLicenseClass(tt.class)
			if result != tt.valid {
				t.Errorf("Expected %v for license class %s but got %v", tt.valid, tt.class, result)
			}
		})
	}
}

func TestIsValidThaiIDNumber(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid Thai ID", "1320500269543", true},     // Valid Thai ID with correct checksum
		{"sample ID format", "1234567890123", false}, // Will fail checksum validation
		{"too short", "123456789012", false},
		{"too long", "12345678901234", false},
		{"contains letters", "110170123456a", false},
		{"empty", "", false},
		{"all zeros", "0000000000000", false}, // Will fail checksum validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidThaiIDNumber(tt.id)
			if result != tt.valid {
				t.Errorf("Expected %v for ID %s but got %v", tt.valid, tt.id, result)
			}
		})
	}
}

func TestValidateThaiIDChecksum(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid Thai ID checksum", "1320500269543", true}, // Valid checksum
		{"sample ID 1", "1234567890123", false},           // This will fail real checksum validation
		{"sample ID 2", "1234567890124", false},           // This will also fail
		{"too short", "123456789012", false},
		{"too long", "12345678901234", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateThaiIDChecksum(tt.id)
			if result != tt.valid {
				t.Errorf("Expected %v for ID checksum %s but got %v", tt.valid, tt.id, result)
			}
		})
	}
}

func TestGetCountryOrDefault(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "Thailand"},
		{"USA", "USA"},
		{"Japan", "Japan"},
		{"  ", "  "}, // Non-empty but whitespace
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := getCountryOrDefault(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s for input %s but got %s", tt.expected, tt.input, result)
			}
		})
	}
}

func TestAddressIsInBangkok(t *testing.T) {
	tests := []struct {
		name     string
		province string
		expected bool
	}{
		{"Thai Bangkok", "กรุงเทพมหานคร", true},
		{"English Bangkok", "Bangkok", true},
		{"English Bangkok lowercase", "bangkok", true},
		{"Other province", "Chiang Mai", false},
		{"Empty province", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty province test since it won't build successfully
			if tt.province == "" {
				t.Skip("Skipping empty province test - AddressBuilder requires province")
				return
			}

			address, err := NewAddressBuilder().
				WithHouseNumber("123").
				WithSubdistrict("Test").
				WithDistrict("Test").
				WithProvince(tt.province).
				WithPostalCode("10110").
				Build()

			if err != nil {
				t.Errorf("Failed to build address: %v", err)
				return
			}

			result := address.IsInBangkok()
			if result != tt.expected {
				t.Errorf("Expected %v for province %s but got %v", tt.expected, tt.province, result)
			}
		})
	}
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "test@example.com", true},
		{"email with numbers", "user123@domain456.com", true},
		{"email with dots", "first.last@example.com", true},
		{"email with plus", "user+tag@example.com", true},
		{"email with underscore", "user_name@example.com", true},
		{"email with hyphen in domain", "user@sub-domain.com", true},
		{"multiple subdomains", "user@mail.sub.example.com", true},
		{"short domain extension", "user@example.co", true},
		{"long domain extension", "user@example.museum", true},
		{"missing @", "userexample.com", false},
		{"missing domain", "user@", false},
		{"missing local part", "@example.com", false},
		{"double @", "user@@example.com", false},
		{"spaces", "user @example.com", false},
		{"invalid characters", "user$@example.com", false},
		{"domain without extension", "user@domain", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if tt.valid && err != nil {
				t.Errorf("Expected no error for valid email %s but got %v", tt.email, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected error for invalid email %s but got none", tt.email)
			}
		})
	}
}

func TestNewPhonenumber(t *testing.T) {
	tests := []struct {
		name        string
		phonenumber string
		expectClean string
		expectError bool
	}{
		{
			name:        "valid phonenumber start with zero",
			phonenumber: "0890747762",
			expectClean: "0890747762",
			expectError: false,
		},
		{
			name:        "valid phonenumber start with +66",
			phonenumber: "+66890747762",
			expectClean: "+66890747762",
			expectError: false,
		},
		{
			name:        "valid phonenumber with special character",
			phonenumber: "(+66)89-0747-762",
			expectClean: "+66890747762",
			expectError: false,
		},
		{
			name:        "empty phonnumber",
			phonenumber: "",
			expectError: true,
		},
		{
			name:        "invalid phonenumber with less than 10 digits and start with 0",
			phonenumber: "089074762",
			expectClean: "",
			expectError: true,
		},
		{
			name:        "invalid phonenumber with less than 10 digits and start with +66",
			phonenumber: "+668907476",
			expectClean: "",
			expectError: true,
		},
		{
			name:        "invalid phonenumber with non digit character and start with +66",
			phonenumber: "+6689074762a",
			expectClean: "",
			expectError: true,
		},
		{
			name:        "invalid phonenumber not start with 0",
			phonenumber: "6689074776",
			expectClean: "",
			expectError: true,
		},
		{
			name:        "invalid phonenumber with non digit character and start with 0",
			phonenumber: "089074762a",
			expectClean: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phonenumber, err := NewPhonenumber(tt.phonenumber)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if phonenumber != nil {
					t.Errorf("Expected nil email but got %v", phonenumber)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
					return
				}
				if phonenumber == nil {
					t.Errorf("Expected email but got nil")
					return
				}
				if phonenumber.String() != strings.ToLower(tt.expectClean) {
					t.Errorf("Expected %s but got %s", strings.ToLower(tt.phonenumber), phonenumber.String())
				}
			}
		})
	}
}
