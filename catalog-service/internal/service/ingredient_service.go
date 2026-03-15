package service

import (
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

type IngredientService interface {
	GetAllIngredients() ([]model.Ingredient, error)
	CreateIngredient(ingredient *model.Ingredient) error
}

type ingredientService struct {
	repo repository.IngredientRepository
}

func NewIngredientService(repo repository.IngredientRepository) IngredientService {
	return &ingredientService{repo: repo}
}

func (s *ingredientService) GetAllIngredients() ([]model.Ingredient, error) {
	return s.repo.FindAll()
}

func (s *ingredientService) CreateIngredient(ingredient *model.Ingredient) error {
	return s.repo.Create(ingredient)
}
