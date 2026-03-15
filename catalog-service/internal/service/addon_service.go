package service

import (
	"errors"

	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

type AddOnService interface {
	GetAddOnsByMenuItemID(menuItemID uint) ([]model.MenuAddOn, error)
	CreateAddOn(addon *model.MenuAddOn) error
	DeleteAddOn(id uint) error
}

type addOnService struct {
	repo           repository.AddOnRepository
	ingredientRepo repository.IngredientRepository
}

func NewAddOnService(repo repository.AddOnRepository, ingredientRepo repository.IngredientRepository) AddOnService {
	return &addOnService{repo: repo, ingredientRepo: ingredientRepo}
}

func (s *addOnService) GetAddOnsByMenuItemID(menuItemID uint) ([]model.MenuAddOn, error) {
	return s.repo.FindByMenuItemID(menuItemID)
}

func (s *addOnService) CreateAddOn(addon *model.MenuAddOn) error {
	if _, err := s.ingredientRepo.FindByID(addon.IngredientID); err != nil {
		return errors.New("ingredient not found")
	}
	return s.repo.Create(addon)
}

func (s *addOnService) DeleteAddOn(id uint) error {
	return s.repo.Delete(id)
}
