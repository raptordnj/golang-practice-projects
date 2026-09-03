package repositories

import (
	"errors"
	"go-mvc/models"

	"github.com/beego/beego/v2/client/orm"
)

// UserRepository defines the database operations for users
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
}

type userRepository struct {
	o orm.Ormer
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository() UserRepository {
	return &userRepository{
		o: orm.NewOrm(),
	}
}

// Create inserts a new user record
func (r *userRepository) Create(user *models.User) error {
	_, err := r.o.Insert(user)
	return err
}

// FindByEmail retrieves a user by email address
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{Email: email}
	err := r.o.Read(user, "Email")
	if errors.Is(err, orm.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}

// FindByID retrieves a user by ID
func (r *userRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{Id: id}
	err := r.o.Read(user)
	if errors.Is(err, orm.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}
