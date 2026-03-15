package service

import (
	"errors"

	"inventory-service/internal/model"
	"inventory-service/internal/repository"
)

var ErrMaterialNotFound = errors.New("raw material not found")
var ErrMaterialNameExists = errors.New("raw material name already exists")

type MaterialService interface {
	GetAll() ([]model.RawMaterial, error)
	GetByID(id uint) (*model.RawMaterial, error)
	GetLowStock() ([]model.RawMaterial, error)
	Create(m *model.RawMaterial) error
	Update(id uint, input *model.RawMaterial) (*model.RawMaterial, error)
}

type materialService struct {
	repo repository.MaterialRepository
}

func NewMaterialService(repo repository.MaterialRepository) MaterialService {
	return &materialService{repo: repo}
}

func (s *materialService) GetAll() ([]model.RawMaterial, error) {
	return s.repo.FindAll()
}

func (s *materialService) GetByID(id uint) (*model.RawMaterial, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrMaterialNotFound
	}
	return m, nil
}

func (s *materialService) GetLowStock() ([]model.RawMaterial, error) {
	return s.repo.FindLowStock()
}

func (s *materialService) Create(m *model.RawMaterial) error {
	return s.repo.Create(m)
}

func (s *materialService) Update(id uint, input *model.RawMaterial) (*model.RawMaterial, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrMaterialNotFound
	}
	existing.Name = input.Name
	existing.Unit = input.Unit
	existing.ReorderPoint = input.ReorderPoint
	existing.CatalogIngredientID = input.CatalogIngredientID
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}
