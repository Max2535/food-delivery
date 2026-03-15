package service

import (
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

type BOMService interface {
	GetBOMByMenuItemID(menuItemID uint) ([]model.BOMItem, error)
	AddBOMItem(menuItemID uint, ingredientID uint, quantity float64) (*model.BOMItem, error)
	DeleteBOMItem(id uint) error
}

type bomService struct {
	bomRepo        repository.BOMRepository
	ingredientRepo repository.IngredientRepository
}

func NewBOMService(bomRepo repository.BOMRepository, ingredientRepo repository.IngredientRepository) BOMService {
	return &bomService{bomRepo: bomRepo, ingredientRepo: ingredientRepo}
}

func (s *bomService) GetBOMByMenuItemID(menuItemID uint) ([]model.BOMItem, error) {
	return s.bomRepo.FindByMenuItemID(menuItemID)
}

func (s *bomService) AddBOMItem(menuItemID uint, ingredientID uint, quantity float64) (*model.BOMItem, error) {
	if _, err := s.ingredientRepo.FindByID(ingredientID); err != nil {
		return nil, err
	}
	item := &model.BOMItem{
		MenuItemID:   menuItemID,
		IngredientID: ingredientID,
		Quantity:     quantity,
	}
	if err := s.bomRepo.AddItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *bomService) DeleteBOMItem(id uint) error {
	return s.bomRepo.DeleteItem(id)
}
