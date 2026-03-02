package unitofwork

import "github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"

type UnitOfWork interface {
	GetCustomerProfileRepository() customerprofile.Repository
}
