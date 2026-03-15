package service

import (
	"errors"

	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

var ErrChoiceGroupNotFound = errors.New("choice group not found")
var ErrInvalidChoiceRange = errors.New("min_choices must be >= 1 and <= max_choices")

type ChoiceService interface {
	GetChoicesByMenuItemID(menuItemID uint) ([]model.BOMChoiceGroup, error)
	CreateChoiceGroup(group *model.BOMChoiceGroup) error
	DeleteChoiceGroup(id uint) error
	AddChoiceOption(groupID uint, ingredientID uint, quantity float64, extraPrice float64) (*model.BOMChoiceOption, error)
	DeleteChoiceOption(id uint) error
}

type choiceService struct {
	repo           repository.ChoiceRepository
	ingredientRepo repository.IngredientRepository
}

func NewChoiceService(repo repository.ChoiceRepository, ingredientRepo repository.IngredientRepository) ChoiceService {
	return &choiceService{repo: repo, ingredientRepo: ingredientRepo}
}

func (s *choiceService) GetChoicesByMenuItemID(menuItemID uint) ([]model.BOMChoiceGroup, error) {
	return s.repo.FindGroupsByMenuItemID(menuItemID)
}

func (s *choiceService) CreateChoiceGroup(group *model.BOMChoiceGroup) error {
	if group.MinChoices < 1 || group.MaxChoices < group.MinChoices {
		return ErrInvalidChoiceRange
	}
	return s.repo.CreateGroup(group)
}

func (s *choiceService) DeleteChoiceGroup(id uint) error {
	if _, err := s.repo.FindGroupByID(id); err != nil {
		return ErrChoiceGroupNotFound
	}
	return s.repo.DeleteGroup(id)
}

func (s *choiceService) AddChoiceOption(groupID uint, ingredientID uint, quantity float64, extraPrice float64) (*model.BOMChoiceOption, error) {
	if _, err := s.repo.FindGroupByID(groupID); err != nil {
		return nil, ErrChoiceGroupNotFound
	}
	if _, err := s.ingredientRepo.FindByID(ingredientID); err != nil {
		return nil, errors.New("ingredient not found")
	}
	option := &model.BOMChoiceOption{
		GroupID:      groupID,
		IngredientID: ingredientID,
		Quantity:     quantity,
		ExtraPrice:   extraPrice,
	}
	if err := s.repo.AddOption(option); err != nil {
		return nil, err
	}
	return option, nil
}

func (s *choiceService) DeleteChoiceOption(id uint) error {
	return s.repo.DeleteOption(id)
}
