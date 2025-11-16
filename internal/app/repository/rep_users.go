package repository

import (
	"errors"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"

	"gorm.io/gorm"
)

func (r *Repository) RegisterUser(login, password string) error {
	u := model.Users{Login: login, Password: password}
	return r.db.Create(&u).Error
}

func (r *Repository) GetUserByID(id uint) (*model.Users, error) {
	var u model.Users
	if err := r.db.First(&u, "id_user = ?", id).Error; err != nil {
		return nil, err
	}
	u.Password = ""
	return &u, nil
}

func (r *Repository) GetUserByUsername(login string) (*model.Users, error) {
	var u model.Users
	if err := r.db.First(&u, "login = ?", login).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateUser(id uint, password *string) (*model.Users, error) {
	var u model.Users
	if err := r.db.First(&u, "id_user = ?", id).Error; err != nil {
		return nil, err
	}
	if password != nil && *password != "" {
		u.Password = *password
	}
	if err := r.db.Save(&u).Error; err != nil {
		return nil, err
	}
	u.Password = ""
	return &u, nil
}

func (r *Repository) CheckUser(login, password string) error {
	var u model.Users
	if err := r.db.Where("login = ?", login).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid credentials")
		}
		return err
	}
	if u.Password != password {
		return errors.New("invalid credentials")
	}
	return nil
}
