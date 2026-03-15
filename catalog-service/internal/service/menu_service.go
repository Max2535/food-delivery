package service

import (
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
	"errors"
)

var ErrMenuItemAlreadyExists = errors.New("menu item name already exists")

type MenuService interface {
	GetAllMenuItems() ([]model.MenuItem, error)
	GetMenuItemByID(id uint) (*model.MenuItem, error)
	CreateMenuItem(item *model.MenuItem) error
	UpdateMenuItem(id uint, item *model.MenuItem) (*model.MenuItem, error)
	DeleteMenuItem(id uint) error
}

type menuService struct {
	repo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{repo: repo}
}

func (s *menuService) GetAllMenuItems() ([]model.MenuItem, error) {
	return s.repo.FindAll()
}

func (s *menuService) GetMenuItemByID(id uint) (*model.MenuItem, error) {
	return s.repo.FindByID(id)
}

func (s *menuService) CreateMenuItem(item *model.MenuItem) error {
	existing, _ := s.repo.FindByName(item.Name)
	if existing != nil {
		return ErrMenuItemAlreadyExists
	}
	return s.repo.Create(item)
}

func (s *menuService) UpdateMenuItem(id uint, input *model.MenuItem) (*model.MenuItem, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if name is changing and if new name already exists
	if existing.Name != input.Name {
		conflict, _ := s.repo.FindByName(input.Name)
		if conflict != nil {
			return nil, ErrMenuItemAlreadyExists
		}
	}

	existing.Name = input.Name
	existing.Description = input.Description
	existing.Price = input.Price
	existing.Category = input.Category
	existing.IsAvailable = input.IsAvailable
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *menuService) DeleteMenuItem(id uint) error {
	return s.repo.Delete(id)
}
