package service

import (
	"errors"

	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

var ErrInvalidMultiplier = errors.New("quantity_multiplier must be greater than 0")

type PortionService interface {
	GetPortionsByMenuItemID(menuItemID uint) ([]model.MenuPortionSize, error)
	CreatePortion(portion *model.MenuPortionSize) error
	DeletePortion(id uint) error
}

type portionService struct {
	repo     repository.PortionRepository
	menuRepo repository.MenuRepository
}

func NewPortionService(repo repository.PortionRepository, menuRepo repository.MenuRepository) PortionService {
	return &portionService{repo: repo, menuRepo: menuRepo}
}

func (s *portionService) GetPortionsByMenuItemID(menuItemID uint) ([]model.MenuPortionSize, error) {
	return s.repo.FindByMenuItemID(menuItemID)
}

func (s *portionService) CreatePortion(portion *model.MenuPortionSize) error {
	if portion.QuantityMultiplier <= 0 {
		return ErrInvalidMultiplier
	}
	if _, err := s.menuRepo.FindByID(portion.MenuItemID); err != nil {
		return errors.New("menu item not found")
	}
	return s.repo.Create(portion)
}

func (s *portionService) DeletePortion(id uint) error {
	return s.repo.Delete(id)
}
