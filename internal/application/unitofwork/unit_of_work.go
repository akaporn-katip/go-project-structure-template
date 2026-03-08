package unitofwork

import "github.com/akaporn-katip/go-project-structure-template/internal/application/repositories"

type UnitOfWork interface {
	GetRepositories() repositories.Repositories
}
