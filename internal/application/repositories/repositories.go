package repositories

import "github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"

type Repositories interface {
	GetCustomerProfileRepository() customerprofile.Repository
}
