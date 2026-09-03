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
func NewUserRepository(o ...orm.Ormer) UserRepository {
	var ormer orm.Ormer
	if len(o) > 0 {
		ormer = o[0]
	}
	return &userRepository{
		o: ormer,
	}
}

func (r *userRepository) orm() orm.Ormer {
	if r.o != nil {
		return r.o
	}
	return orm.NewOrm()
}

// Create inserts a new user record
func (r *userRepository) Create(user *models.User) error {
	_, err := r.orm().Insert(user)
	return err
}

// FindByEmail retrieves a user by email address
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{Email: email}
	err := r.orm().Read(user, "Email")
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// FindByID retrieves a user by ID
func (r *userRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{Id: id}
	err := r.orm().Read(user)
	if err != nil {
		if errors.Is(err, orm.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}
