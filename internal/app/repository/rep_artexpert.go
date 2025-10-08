package repository

import (
	"fmt"

	"github.com/w1zZzyy22/art-analysis/internal/app/model"
)

func (r *Repository) GetExperts() ([]model.ArtExpert, error) {
	var experts []model.ArtExpert

	err := r.db.Find(&experts).Error
	if err != nil {
		return nil, err
	}

	if len(experts) == 0 {
		return nil, fmt.Errorf("experts not found")
	}
	return experts, nil
}

func (r *Repository) GetExpertsByName(title string) ([]model.ArtExpert, error) {
	var experts []model.ArtExpert
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&experts).Error
	if err != nil {
		return nil, err
	}
	return experts, nil
}

func (r *Repository) GetExpertByID(id int) (*model.ArtExpert, error) {
	var expert model.ArtExpert
	err := r.db.First(&expert, id).Error
	if err != nil {
		return nil, err
	}
	return &expert, nil
}
