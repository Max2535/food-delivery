package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type IngredientRepository interface {
	FindAll() ([]model.Ingredient, error)
	FindByID(id uint) (*model.Ingredient, error)
	Create(ingredient *model.Ingredient) error
}

type ingredientRepository struct {
	db *gorm.DB
}

func NewIngredientRepository(db *gorm.DB) IngredientRepository {
	return &ingredientRepository{db: db}
}

func (r *ingredientRepository) FindAll() ([]model.Ingredient, error) {
	var ingredients []model.Ingredient
	err := r.db.Find(&ingredients).Error
	return ingredients, err
}

func (r *ingredientRepository) FindByID(id uint) (*model.Ingredient, error) {
	var ingredient model.Ingredient
	err := r.db.First(&ingredient, id).Error
	if err != nil {
		return nil, err
	}
	return &ingredient, nil
}

func (r *ingredientRepository) Create(ingredient *model.Ingredient) error {
	return r.db.Create(ingredient).Error
}
