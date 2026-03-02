
## Project Structure
```
crm-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   ├── errors.go
│   │   │   └── value_objects.go
│   │   ├── product/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── errors.go
│   │   └── order/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── events.go
│   ├── application/
│   │   ├── user/
│   │   │   ├── command/
│   │   │   │   ├── create_user.go
│   │   │   │   ├── update_user.go
│   │   │   │   └── delete_user.go
│   │   │   ├── query/
│   │   │   │   ├── get_user.go
│   │   │   │   └── list_users.go
│   │   │   └── dto/
│   │   │       ├── user_request.go
│   │   │       └── user_response.go
│   │   ├── product/
│   │   │   └── ...
│   │   └── order/
│   │       └── ...
│   ├── infrastructure/
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── product_handler.go
│   │   │   │   └── order_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── logging.go
│   │   │   │   ├── cors.go
│   │   │   │   └── error_handler.go
│   │   │   ├── router.go
│   │   │   └── response.go
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── product_repository.go
│   │   │   │   ├── order_repository.go
│   │   │   │   └── migrations/
│   │   │   │       ├── 001_create_users.sql
│   │   │   │       ├── 002_create_products.sql
│   │   │   │       └── 003_create_orders.sql
│   │   │   └── redis/
│   │   │       └── cache_repository.go
│   │   ├── messaging/
│   │   │   └── event_publisher.go
│   │   └── config/
│   │       └── database.go
│   └── shared/
│       ├── errors/
│       │   └── errors.go
│       ├── logger/
│       │   └── logger.go
│       └── validator/
│           └── validator.go
├── pkg/
│   ├── contextutil/
│   │   └── context.go
│   └── pagination/
│       └── pagination.go
├── config/
│   ├── config.go
│   └── config.yaml
├── api/
│   └── openapi.yaml
├── scripts/
│   ├── setup.sh
│   └── migrate.sh
├── deployments/
│   └── docker/
│       ├── Dockerfile
│       └── docker-compose.yml
├── tests/
│   ├── integration/
│   │   └── user_test.go
│   └── fixtures/
│       └── users.json
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── .gitignore
└── README.md
```
## 1. **Constructor Struct Pattern** (Recommended)

````go
package customerprofile

type CustomerProfileInput struct {
	Firstname   string
	Lastname    string
	DateOfBirth string
	Email       string
}

func NewCustomer(input CustomerProfileInput) (*CustomerProfile, error) {
	dof, err := NewDateOfBirth(input.DateOfBirth)
	if err != nil {
		return nil, err
	}

	email, err := NewEmail(input.Email)
	if err != nil {
		return nil, err
	}

	return &CustomerProfile{
		Firstname:   input.Firstname,
		Lastname:    input.Lastname,
		Email:       *email,
		DateOfBirth: *dof,
	}, nil
}
````

## 2. **Builder Pattern**

````go
type CustomerProfileBuilder struct {
	firstname   string
	lastname    string
	dateOfBirth string
	email       string
	address     Address
}

func NewCustomerBuilder() *CustomerProfileBuilder {
	return &CustomerProfileBuilder{}
}

func (b *CustomerProfileBuilder) WithName(firstname, lastname string) *CustomerProfileBuilder {
	b.firstname = firstname
	b.lastname = lastname
	return b
}

func (b *CustomerProfileBuilder) WithDateOfBirth(dob string) *CustomerProfileBuilder {
	b.dateOfBirth = dob
	return b
}

func (b *CustomerProfileBuilder) WithEmail(email string) *CustomerProfileBuilder {
	b.email = email
	return b
}

func (b *CustomerProfileBuilder) WithAddress(address Address) *CustomerProfileBuilder {
	b.address = address
	return b
}

func (b *CustomerProfileBuilder) Build() (*CustomerProfile, error) {
	dof, err := NewDateOfBirth(b.dateOfBirth)
	if err != nil {
		return nil, err
	}

	email, err := NewEmail(b.email)
	if err != nil {
		return nil, err
	}

	return &CustomerProfile{
		Firstname:      b.firstname,
		Lastname:       b.lastname,
		Email:          *email,
		DateOfBirth:    *dof,
		CurrentAddress: b.address,
	}, nil
}
````

## 3. **Functional Options Pattern**

````go
type CustomerOption func(*CustomerProfile) error

func WithName(firstname, lastname string) CustomerOption {
	return func(c *CustomerProfile) error {
		c.Firstname = firstname
		c.Lastname = lastname
		return nil
	}
}

func WithDateOfBirth(dob string) CustomerOption {
	return func(c *CustomerProfile) error {
		dateOfBirth, err := NewDateOfBirth(dob)
		if err != nil {
			return err
		}
		c.DateOfBirth = *dateOfBirth
		return nil
	}
}

func WithEmail(email string) CustomerOption {
	return func(c *CustomerProfile) error {
		emailVO, err := NewEmail(email)
		if err != nil {
			return err
		}
		c.Email = *emailVO
		return nil
	}
}

func NewCustomer(opts ...CustomerOption) (*CustomerProfile, error) {
	customer := &CustomerProfile{}
	
	for _, opt := range opts {
		if err := opt(customer); err != nil {
			return nil, err
		}
	}
	
	return customer, nil
}
````

## Usage Examples:

**Constructor Struct:**
```go
customer, err := NewCustomer(CustomerProfileInput{
	Firstname:   "John",
	Lastname:    "Doe",
	DateOfBirth: "1990-01-01",
	Email:       "john@example.com",
})
```

**Builder Pattern:**
```go
customer, err := NewCustomerBuilder().
	WithName("John", "Doe").
	WithDateOfBirth("1990-01-01").
	WithEmail("john@example.com").
	Build()
```

**Functional Options:**
```go
customer, err := NewCustomer(
	WithName("John", "Doe"),
	WithDateOfBirth("1990-01-01"),
	WithEmail("john@example.com"),
)
```

## Recommendation:

For your use case, I recommend the **Constructor Struct Pattern** because:
- Simple and clear
- Easy to validate all inputs at once
- Good IDE support with autocompletion
- Minimal boilerplate code
- Easy to extend with new fields

The Builder pattern is better when you have complex object construction with many optional fields, while Functional Options are great for highly configurable APIs.
