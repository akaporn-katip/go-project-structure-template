package customerprofile

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Address struct {
	houseNumber string // house number
	building    string // building name/number
	moo         string // village number (หมู่)
	soi         string // alley (ซอย)
	road        string // road (ถนน)
	subdistrict string // tambon / khwaeng (ตำบล / แขวง)
	district    string // amphoe / khet (อำเภอ / เขต)
	province    string // changwat (จังหวัด)
	postalCode  string
	country     string
}

type AddressBuilder struct {
	houseNumber string // house number
	building    string // building name/number
	moo         string // village number (หมู่)
	soi         string // alley (ซอย)
	road        string // road (ถนน)
	subdistrict string // tambon / khwaeng (ตำบล / แขวง)
	district    string // amphoe / khet (อำเภอ / เขต)
	province    string // changwat (จังหวัด)
	postalCode  string
	country     string
}

func NewAddressBuilder() *AddressBuilder {
	return &AddressBuilder{}
}

func (ab *AddressBuilder) WithHouseNumber(houseNumber string) *AddressBuilder {
	ab.houseNumber = houseNumber
	return ab
}

func (ab *AddressBuilder) WithBuilding(building string) *AddressBuilder {
	ab.building = building
	return ab
}

func (ab *AddressBuilder) WithMoo(moo string) *AddressBuilder {
	ab.moo = moo
	return ab
}

func (ab *AddressBuilder) WithSoi(soi string) *AddressBuilder {
	ab.soi = soi
	return ab
}

func (ab *AddressBuilder) WithRoad(road string) *AddressBuilder {
	ab.road = road
	return ab
}

func (ab *AddressBuilder) WithSubdistrict(subdistrict string) *AddressBuilder {
	ab.subdistrict = subdistrict
	return ab
}

func (ab *AddressBuilder) WithDistrict(district string) *AddressBuilder {
	ab.district = district
	return ab
}

func (ab *AddressBuilder) WithProvince(province string) *AddressBuilder {
	ab.province = province
	return ab
}

func (ab *AddressBuilder) WithPostalCode(postalCode string) *AddressBuilder {
	ab.postalCode = postalCode
	return ab
}

func (ab *AddressBuilder) WithCountry(country string) *AddressBuilder {
	ab.country = country
	return ab
}

func (ab *AddressBuilder) Build() (*Address, error) {
	// Validation
	if ab.houseNumber == "" {
		return nil, ErrInvalidAddressHouseNumber
	}
	if ab.subdistrict == "" {
		return nil, ErrInvalidAddressSubdistrict
	}
	if ab.district == "" {
		return nil, ErrInvalidAddressDistrict
	}
	if ab.province == "" {
		return nil, ErrInvalidAddressProvince
	}
	if ab.postalCode == "" {
		return nil, ErrInvalidAddressPostalCode
	}
	if !isValidThaiPostalCode(ab.postalCode) {
		return nil, ErrInvalidThaiPostalCode
	}

	return &Address{
		houseNumber: ab.houseNumber,
		building:    ab.building,
		moo:         ab.moo,
		soi:         ab.soi,
		road:        ab.road,
		subdistrict: ab.subdistrict,
		district:    ab.district,
		province:    ab.province,
		postalCode:  ab.postalCode,
		country:     getCountryOrDefault(ab.country),
	}, nil
}

// Getter methods
func (a Address) HouseNumber() string {
	return a.houseNumber
}

func (a Address) Building() string {
	return a.building
}

func (a Address) Moo() string {
	return a.moo
}

func (a Address) Soi() string {
	return a.soi
}

func (a Address) Road() string {
	return a.road
}

func (a Address) Subdistrict() string {
	return a.subdistrict
}

func (a Address) District() string {
	return a.district
}

func (a Address) Province() string {
	return a.province
}

func (a Address) PostalCode() string {
	return a.postalCode
}

func (a Address) Country() string {
	return a.country
}

// Business methods
func (a Address) FullAddress() string {
	parts := []string{}

	if a.houseNumber != "" {
		parts = append(parts, a.houseNumber)
	}
	if a.building != "" {
		parts = append(parts, a.building)
	}
	if a.moo != "" {
		parts = append(parts, "หมู่ "+a.moo)
	}
	if a.soi != "" {
		parts = append(parts, "ซอย "+a.soi)
	}
	if a.road != "" {
		parts = append(parts, "ถนน "+a.road)
	}

	parts = append(parts, a.subdistrict, a.district, a.province, a.postalCode)

	if a.country != "" && a.country != "Thailand" {
		parts = append(parts, a.country)
	}

	return strings.Join(parts, " ")
}

func (a Address) IsInBangkok() bool {
	return strings.ToLower(a.province) == "กรุงเทพมหานคร" ||
		strings.ToLower(a.province) == "bangkok"
}

// Helper functions
func isValidThaiPostalCode(code string) bool {
	if len(code) != 5 {
		return false
	}

	for _, r := range code {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func getCountryOrDefault(country string) string {
	if country == "" {
		return "Thailand"
	}
	return country
}

type IdentityCard struct {
	number      string
	title       string
	firstname   string
	lastname    string
	dateOfBirth DateOfBirth
	issueDate   time.Time
	expiryDate  time.Time
	address     Address
}

type IdentityCardInput struct {
	Number      string
	Title       string
	Firstname   string
	Lastname    string
	DateOfBirth string
	IssueDate   string
	ExpiryDate  string
	Address     Address
}

func NewIdentityCard(input IdentityCardInput) (*IdentityCard, error) {
	// Validate ID number (Thai ID format: 13 digits)
	if !isValidThaiIDNumber(input.Number) {
		return nil, ErrInvalidThaiIDNumber
	}

	// Validate names
	if strings.TrimSpace(input.Firstname) == "" {
		return nil, ErrInvalidFirstname
	}
	if strings.TrimSpace(input.Lastname) == "" {
		return nil, ErrInvalidLastname
	}

	// Parse date of birth
	dob, err := NewDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}

	// Parse dates
	issueDate, parseErr := time.Parse("2006-01-02", input.IssueDate)
	if parseErr != nil {
		return nil, ErrInvalidIssueDate
	}

	expiryDate, parseErr := time.Parse("2006-01-02", input.ExpiryDate)
	if parseErr != nil {
		return nil, ErrInvalidExpiryDate
	}

	// Validate date logic
	if issueDate.After(expiryDate) {
		return nil, ErrIssueDateAfterExpiry
	}

	return &IdentityCard{
		number:      input.Number,
		title:       strings.TrimSpace(input.Title),
		firstname:   strings.TrimSpace(input.Firstname),
		lastname:    strings.TrimSpace(input.Lastname),
		dateOfBirth: *dob,
		issueDate:   issueDate,
		expiryDate:  expiryDate,
		address:     input.Address,
	}, nil
}

func (id IdentityCard) ID() string {
	return id.number
}

func isValidThaiIDNumber(id string) bool {
	// Thai ID must be 13 digits
	if len(id) != 13 {
		return false
	}

	// Check if all characters are digits
	for _, char := range id {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Validate Thai ID checksum algorithm
	return validateThaiIDChecksum(id)
}

func validateThaiIDChecksum(id string) bool {
	if len(id) != 13 {
		return false
	}

	// Thai ID checksum calculation
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(id[i] - '0')
		sum += digit * (13 - i)
	}

	checksum := (11 - (sum % 11)) % 10
	lastDigit := int(id[12] - '0')

	return checksum == lastDigit
}

type Passport struct {
	number         string
	issuedDate     time.Time
	expiryDate     time.Time
	issuingCountry string
}

type PassportInput struct {
	Number         string
	IssuedDate     string
	ExpiryDate     string
	IssuingCountry string
}

func NewPassport(input PassportInput) (*Passport, error) {
	// Validate passport number
	if strings.TrimSpace(input.Number) == "" {
		return nil, ErrInvalidPassportNumber
	}

	// Validate issuing country
	if strings.TrimSpace(input.IssuingCountry) == "" {
		return nil, ErrInvalidIssuingCountry
	}

	// Parse dates
	issuedDate, parseErr := time.Parse("2006-01-02", input.IssuedDate)
	if parseErr != nil {
		return nil, ErrInvalidPassportIssueDate
	}

	expiryDate, parseErr := time.Parse("2006-01-02", input.ExpiryDate)
	if parseErr != nil {
		return nil, ErrInvalidPassportExpiryDate
	}

	// Validate date logic
	if issuedDate.After(expiryDate) {
		return nil, ErrPassportIssueDateAfterExpiry
	}

	// Validate passport is not too old or too far in future
	now := time.Now()
	if issuedDate.After(now) {
		return nil, ErrPassportFutureIssueDate
	}
	if expiryDate.Before(now.AddDate(-20, 0, 0)) { // 20 years old
		return nil, ErrPassportTooOld
	}

	return &Passport{
		number:         strings.TrimSpace(strings.ToUpper(input.Number)),
		issuedDate:     issuedDate,
		expiryDate:     expiryDate,
		issuingCountry: strings.TrimSpace(input.IssuingCountry),
	}, nil
}

// Getter methods
func (p Passport) Number() string {
	return p.number
}

func (p Passport) IssuedDate() time.Time {
	return p.issuedDate
}

func (p Passport) ExpiryDate() time.Time {
	return p.expiryDate
}

func (p Passport) IssuingCountry() string {
	return p.issuingCountry
}

// Business methods
func (p Passport) IsExpired() bool {
	return time.Now().After(p.expiryDate)
}

func (p Passport) IsValid() bool {
	return !p.IsExpired() && p.number != ""
}

func (p Passport) DaysUntilExpiry() int {
	if p.IsExpired() {
		return 0
	}
	return int(time.Until(p.expiryDate).Hours() / 24)
}

func (p Passport) IsExpiringSoon(days int) bool {
	if p.IsExpired() {
		return true
	}
	return p.DaysUntilExpiry() <= days
}

func (p Passport) ValidityPeriodInYears() int {
	return int(p.expiryDate.Sub(p.issuedDate).Hours() / (24 * 365))
}

func (p Passport) IsThaiPassport() bool {
	return strings.ToLower(p.issuingCountry) == "thailand" ||
		strings.ToLower(p.issuingCountry) == "thai"
}

func (p Passport) String() string {
	return fmt.Sprintf("Passport %s (%s) - Valid until %s",
		p.number,
		p.issuingCountry,
		p.expiryDate.Format("2006-01-02"))
}

type DrivingLicense struct {
	number           string
	issuedDate       time.Time
	expiryDate       time.Time
	class            string
	issuingAuthority string
}

type DrivingLicenseInput struct {
	Number           string
	IssuedDate       string
	ExpiryDate       string
	Class            string
	IssuingAuthority string
}

func NewDrivingLicense(input DrivingLicenseInput) (*DrivingLicense, error) {
	// Validate license number
	if strings.TrimSpace(input.Number) == "" {
		return nil, ErrInvalidDrivingLicenseNumber
	}

	// Validate license class
	if !isValidThaiLicenseClass(input.Class) {
		return nil, ErrInvalidDrivingLicenseClass
	}

	// Validate issuing authority
	if strings.TrimSpace(input.IssuingAuthority) == "" {
		return nil, ErrInvalidIssuingAuthority
	}

	// Parse dates
	issuedDate, parseErr := time.Parse("2006-01-02", input.IssuedDate)
	if parseErr != nil {
		return nil, ErrInvalidDrivingLicenseIssueDate
	}

	expiryDate, parseErr := time.Parse("2006-01-02", input.ExpiryDate)
	if parseErr != nil {
		return nil, ErrInvalidDrivingLicenseExpiryDate
	}

	// Validate date logic
	if issuedDate.After(expiryDate) {
		return nil, ErrDrivingLicenseIssueDateAfterExpiry
	}

	// Validate dates are reasonable
	now := time.Now()
	if issuedDate.After(now) {
		return nil, ErrDrivingLicenseFutureIssueDate
	}

	return &DrivingLicense{
		number:           strings.TrimSpace(strings.ToUpper(input.Number)),
		issuedDate:       issuedDate,
		expiryDate:       expiryDate,
		class:            strings.TrimSpace(strings.ToUpper(input.Class)),
		issuingAuthority: strings.TrimSpace(input.IssuingAuthority),
	}, nil
}

// Getter methods
func (dl DrivingLicense) Number() string {
	return dl.number
}

func (dl DrivingLicense) IssuedDate() time.Time {
	return dl.issuedDate
}

func (dl DrivingLicense) ExpiryDate() time.Time {
	return dl.expiryDate
}

func (dl DrivingLicense) Class() string {
	return dl.class
}

func (dl DrivingLicense) IssuingAuthority() string {
	return dl.issuingAuthority
}

func (dl DrivingLicense) IsExpired() bool {
	return time.Now().After(dl.expiryDate)
}

func (dl DrivingLicense) IsValid() bool {
	return !dl.IsExpired() && dl.number != ""
}

func (dl DrivingLicense) DaysUntilExpiry() int {
	if dl.IsExpired() {
		return 0
	}
	return int(time.Until(dl.expiryDate).Hours() / 24)
}

func (dl DrivingLicense) IsExpiringSoon(days int) bool {
	if dl.IsExpired() {
		return true
	}
	return dl.DaysUntilExpiry() <= days
}

func (dl DrivingLicense) CanDriveMotorcycle() bool {
	motorcycleClasses := []string{"A1", "A2", "A3", "A"}
	for _, class := range motorcycleClasses {
		if dl.class == class {
			return true
		}
	}
	return false
}

func (dl DrivingLicense) CanDriveCar() bool {
	carClasses := []string{"B1", "B2", "B3", "B"}
	for _, class := range carClasses {
		if dl.class == class {
			return true
		}
	}
	return false
}

func (dl DrivingLicense) CanDriveTruck() bool {
	truckClasses := []string{"C1", "C2", "C3", "C"}
	for _, class := range truckClasses {
		if dl.class == class {
			return true
		}
	}
	return false
}

func (dl DrivingLicense) GetVehicleType() string {
	switch dl.class {
	case "A1", "A2", "A3", "A":
		return "Motorcycle"
	case "B1", "B2", "B3", "B":
		return "Car"
	case "C1", "C2", "C3", "C":
		return "Truck"
	case "D1", "D2", "D3", "D":
		return "Bus"
	case "E1", "E2", "E3", "E":
		return "Trailer"
	default:
		return "Unknown"
	}
}

func (dl DrivingLicense) String() string {
	return fmt.Sprintf("Driving License %s (Class %s) - Valid until %s",
		dl.number,
		dl.class,
		dl.expiryDate.Format("2006-01-02"))
}

func isValidThaiLicenseClass(class string) bool {
	validClasses := []string{
		"A1", "A2", "A3", "A", // Motorcycles
		"B1", "B2", "B3", "B", // Cars
		"C1", "C2", "C3", "C", // Trucks
		"D1", "D2", "D3", "D", // Buses
		"E1", "E2", "E3", "E", // Trailers
	}

	class = strings.ToUpper(strings.TrimSpace(class))
	for _, validClass := range validClasses {
		if class == validClass {
			return true
		}
	}
	return false
}

type DateOfBirth struct {
	date      time.Time
	precision DatePrecision
}

type DatePrecision int

const (
	YearOnly DatePrecision = iota
	MonthYear
	FullDate
)

func NewDateOfBirth(dob string) (*DateOfBirth, error) {
	// Try different date formats
	formats := []struct {
		layout    string
		precision DatePrecision
	}{
		{"20060102", FullDate}, // 1990-12-25
		{"200601", MonthYear},  // 1990-12
		{"2006", YearOnly},     // 1990
	}

	for _, format := range formats {
		if t, err := time.Parse(format.layout, dob); err == nil {
			// Validate reasonable birth year (not future, not too old)
			currentYear := time.Now().Year()
			birthYear := t.Year()

			if birthYear > currentYear {
				return nil, ErrFutureBirthDate
			}
			if birthYear < (currentYear - 150) {
				return nil, ErrTooOldBirthDate
			}

			return &DateOfBirth{
				date:      t,
				precision: format.precision,
			}, nil
		}
	}

	return nil, ErrInvalidDateOfBirthFormat
}

// Getter methods
func (d DateOfBirth) Date() time.Time {
	return d.date
}

func (d DateOfBirth) Year() int {
	return d.date.Year()
}

func (d DateOfBirth) Month() time.Month {
	if d.precision == YearOnly {
		return 0 // Unknown month
	}
	return d.date.Month()
}

func (d DateOfBirth) Day() int {
	if d.precision != FullDate {
		return 0 // Unknown day
	}
	return d.date.Day()
}

func (d DateOfBirth) Precision() DatePrecision {
	return d.precision
}

func (d DateOfBirth) String() string {
	switch d.precision {
	case YearOnly:
		return fmt.Sprintf("%d", d.date.Year())
	case MonthYear:
		return d.date.Format("200601")
	case FullDate:
		return d.date.Format("20060102")
	default:
		return d.date.String()
	}
}

func (d DateOfBirth) ISOString() string {
	switch d.precision {
	case YearOnly:
		return fmt.Sprintf("%d", d.date.Year())
	case MonthYear:
		return d.date.Format("200601")
	case FullDate:
		return d.date.Format("20060102")
	default:
		return d.date.String()
	}
}

type Email struct {
	value string
}

func (e Email) String() string {
	return e.value
}

func NewEmail(email string) (*Email, error) {
	if err := validateEmail(email); err != nil {
		return nil, ErrInvalidEmailFormat
	}

	return &Email{
		value: strings.ToLower(email),
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// More robust validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	return nil
}

type CustomerID struct {
	value uuid.UUID
}

func NewCustomerID(id string) (*CustomerID, error) {
	if err := uuid.Validate(id); err != nil {
		return nil, ErrInvalidUUIDFormat
	}

	return &CustomerID{
		value: uuid.MustParse(id),
	}, nil
}

func GenerateCustomerID() CustomerID {
	return CustomerID{
		value: uuid.New(),
	}
}

func (c CustomerID) String() string {
	return c.value.String()
}

type PhoneNumber struct {
	number string
}

func NewPhonenumber(number string) (*PhoneNumber, error) {
	if err := validatePhonenumber(number); err != nil {
		return nil, ErrInvalidPhoneNumberFormat
	}
	return &PhoneNumber{
		number: cleanPhonenumber(number),
	}, nil
}

func cleanPhonenumber(number string) string {
	cleaned := strings.ReplaceAll(number, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	return cleaned
}

func validatePhonenumber(number string) error {
	if number == "" {
		return fmt.Errorf("phonenumber cannot be empty")
	}
	// Thai phone numbers are typically 10 digits starting with 0
	// or can be in format +66 followed by 9 digits (without leading 0)
	// Remove common separators for validation
	cleaned := cleanPhonenumber(number)

	// Check for international format +66
	if strings.HasPrefix(cleaned, "+66") {
		cleaned = strings.TrimPrefix(cleaned, "+66")
		if len(cleaned) != 9 {
			return fmt.Errorf("invalid Thai phone number format: must be 9 digits after +66")
		}
		// Verify all digits
		for _, char := range cleaned {
			if char < '0' || char > '9' {
				return fmt.Errorf("invalid Thai phone number format: contains non-digit characters")
			}
		}
		return nil
	}

	// Check for local format (10 digits starting with 0)
	if len(cleaned) != 10 {
		return fmt.Errorf("invalid Thai phone number format: must be 10 digits")
	}

	if !strings.HasPrefix(cleaned, "0") {
		return fmt.Errorf("invalid Thai phone number format: must start with 0")
	}

	// Verify all digits
	for _, char := range cleaned {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid Thai phone number format: contains non-digit characters")
		}
	}

	return nil
}

func (p PhoneNumber) String() string {
	return p.number
}
