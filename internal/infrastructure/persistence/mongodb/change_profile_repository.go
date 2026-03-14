package mongodb

import (
	"context"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/domain/customerprofile"
	"github.com/akaporn-katip/go-project-structure-template/internal/domainerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type CustomerProfileModel struct {
	ID          string    `bson:"_id"`
	Title       string    `bson:"title"`
	FirstName   string    `bson:"first_name"`
	LastName    string    `bson:"last_name"`
	Email       string    `bson:"email"`
	DateOfBirth string    `bson:"date_of_birth"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type CustomerProfileRepository struct {
	customerprofile.Repository
	coll   CollectionExecutor
	tracer trace.Tracer
}

func NewCustomerProfileRepository(coll CollectionExecutor) *CustomerProfileRepository {
	return &CustomerProfileRepository{
		coll:   coll,
		tracer: otel.Tracer("api.katipwork.com/crm/internal/infrastructure/persistence/mongodb/customer_profile_repository"),
	}
}

func (cp *CustomerProfileRepository) Create(ctx context.Context, customerProfile *customerprofile.CustomerProfile) error {
	_, span := cp.tracer.Start(ctx, "CustomerProfileRepository.Create")
	defer span.End()

	doc := CustomerProfileModel{
		ID:          customerProfile.ID().String(),
		Email:       customerProfile.Email().String(),
		FirstName:   customerProfile.Firstname(),
		LastName:    customerProfile.Lastname(),
		Title:       customerProfile.Title(),
		DateOfBirth: customerProfile.DateOfBirth().ISOString(),
		CreatedAt:   customerProfile.CreatedAt(),
		UpdatedAt:   customerProfile.UpdatedAt(),
	}

	_, err := cp.coll.InsertOne(ctx, doc)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cp *CustomerProfileRepository) FindByEmail(ctx context.Context, email customerprofile.Email) (*customerprofile.CustomerProfile, error) {
	_, span := cp.tracer.Start(ctx, "CustomerProfileRepository.FindByEmail")
	defer span.End()

	var customer CustomerProfileModel
	filter := bson.M{"email": email.String()}

	// FindOne returns a *mongo.SingleResult, we call Decode() to map it to our struct
	err := cp.coll.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Ok, err.Error())
			return nil, customerprofile.NewFindByEmailNotFoundError(email.String())
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, domainerrors.NewInternalError("fail to get customer profile by email", err)
	}

	entity, err := customer.toDomain()
	if err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return entity, nil
}

func (cp *CustomerProfileRepository) FindByID(ctx context.Context, id customerprofile.CustomerID) (*customerprofile.CustomerProfile, error) {
	_, span := cp.tracer.Start(ctx, "CustomerProfileRepository.FindByID")
	defer span.End()

	var customer CustomerProfileModel
	filter := bson.M{"_id": id.String()}

	// FindOne returns a *mongo.SingleResult, we call Decode() to map it to our struct
	err := cp.coll.FindOne(ctx, filter).Decode(&customer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			span.SetStatus(codes.Ok, err.Error())
			return nil, customerprofile.NewFindByEmailNotFoundError(id.String())
		}
		span.SetStatus(codes.Error, err.Error())
		return nil, domainerrors.NewInternalError("fail to get customer profile by id", err)
	}

	entity, err := customer.toDomain()
	if err != nil {
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return entity, nil
}

func (customer *CustomerProfileModel) toDomain() (*customerprofile.CustomerProfile, error) {
	return customerprofile.ReconstructCustomer(customer.ID, customerprofile.ReconstructCustomerProps{
		Title:       customer.Title,
		Firstname:   customer.FirstName,
		Lastname:    customer.LastName,
		DateOfBirth: customer.DateOfBirth,
		Email:       customer.Email,
	}, customer.CreatedAt, customer.UpdatedAt)
}
