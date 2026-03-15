package service

import (
	"errors"
	"fmt"

	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

var (
	ErrInvalidBOMEntry  = errors.New("exactly one of ingredient_id or sub_menu_item_id must be set")
	ErrCircularBOM      = errors.New("circular BOM reference detected")
	ErrBOMIngredientNotFound = errors.New("ingredient not found")
	ErrBOMMenuItemNotFound   = errors.New("sub menu item not found")
)

// FlatBOMItem is a fully-resolved raw ingredient entry, used for stock deduction.
// All sub-recipes are recursively expanded and quantities are multiplied through.
type FlatBOMItem struct {
	IngredientID   uint    `json:"ingredient_id"`
	IngredientName string  `json:"ingredient_name"`
	Unit           string  `json:"unit"`
	Quantity       float64 `json:"quantity"`
}

type BOMService interface {
	GetBOMByMenuItemID(menuItemID uint) ([]model.BOMItem, error)
	GetFlatBOM(menuItemID uint) ([]FlatBOMItem, error)
	AddBOMItem(menuItemID uint, ingredientID *uint, subMenuItemID *uint, quantity float64) (*model.BOMItem, error)
	DeleteBOMItem(id uint) error
}

type bomService struct {
	bomRepo        repository.BOMRepository
	ingredientRepo repository.IngredientRepository
	menuRepo       repository.MenuRepository
}

func NewBOMService(bomRepo repository.BOMRepository, ingredientRepo repository.IngredientRepository, menuRepo repository.MenuRepository) BOMService {
	return &bomService{bomRepo: bomRepo, ingredientRepo: ingredientRepo, menuRepo: menuRepo}
}

func (s *bomService) GetBOMByMenuItemID(menuItemID uint) ([]model.BOMItem, error) {
	return s.bomRepo.FindByMenuItemID(menuItemID)
}

// GetFlatBOM recursively expands all sub-recipes and returns a flat list of raw ingredients.
// Quantities are multiplied through every level.
func (s *bomService) GetFlatBOM(menuItemID uint) ([]FlatBOMItem, error) {
	return s.flattenBOM(menuItemID, 1.0, map[uint]bool{})
}

func (s *bomService) flattenBOM(menuItemID uint, multiplier float64, ancestors map[uint]bool) ([]FlatBOMItem, error) {
	if ancestors[menuItemID] {
		return nil, fmt.Errorf("%w: menu item %d", ErrCircularBOM, menuItemID)
	}
	ancestors[menuItemID] = true
	defer func() { delete(ancestors, menuItemID) }()

	items, err := s.bomRepo.FindByMenuItemID(menuItemID)
	if err != nil {
		return nil, err
	}

	var result []FlatBOMItem
	for _, item := range items {
		qty := item.Quantity * multiplier
		if item.SubMenuItemID != nil {
			// Recursively expand sub-recipe; qty acts as a "portion" multiplier
			subItems, err := s.flattenBOM(*item.SubMenuItemID, qty, ancestors)
			if err != nil {
				return nil, err
			}
			result = append(result, subItems...)
		} else if item.IngredientID != nil && item.Ingredient != nil {
			result = append(result, FlatBOMItem{
				IngredientID:   *item.IngredientID,
				IngredientName: item.Ingredient.Name,
				Unit:           item.Ingredient.Unit,
				Quantity:       qty,
			})
		}
	}
	return result, nil
}

func (s *bomService) AddBOMItem(menuItemID uint, ingredientID *uint, subMenuItemID *uint, quantity float64) (*model.BOMItem, error) {
	// Exactly one must be set
	if (ingredientID == nil) == (subMenuItemID == nil) {
		return nil, ErrInvalidBOMEntry
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	if ingredientID != nil {
		if _, err := s.ingredientRepo.FindByID(*ingredientID); err != nil {
			return nil, ErrBOMIngredientNotFound
		}
	}

	if subMenuItemID != nil {
		if _, err := s.menuRepo.FindByID(*subMenuItemID); err != nil {
			return nil, ErrBOMMenuItemNotFound
		}
		// Check that adding this sub-recipe won't create a cycle
		if err := s.checkNoCycle(menuItemID, *subMenuItemID); err != nil {
			return nil, err
		}
	}

	item := &model.BOMItem{
		MenuItemID:    menuItemID,
		IngredientID:  ingredientID,
		SubMenuItemID: subMenuItemID,
		Quantity:      quantity,
	}
	if err := s.bomRepo.AddItem(item); err != nil {
		return nil, err
	}
	return item, nil
}

// checkNoCycle verifies that adding candidateSubID as a sub-recipe of parentID won't create a cycle.
func (s *bomService) checkNoCycle(parentID, candidateSubID uint) error {
	ancestors := map[uint]bool{parentID: true}
	return s.detectCycle(candidateSubID, ancestors)
}

func (s *bomService) detectCycle(menuItemID uint, ancestors map[uint]bool) error {
	if ancestors[menuItemID] {
		return fmt.Errorf("%w: menu item %d", ErrCircularBOM, menuItemID)
	}
	ancestors[menuItemID] = true
	defer func() { delete(ancestors, menuItemID) }()

	items, err := s.bomRepo.FindByMenuItemID(menuItemID)
	if err != nil {
		return nil // non-fatal for cycle check
	}
	for _, item := range items {
		if item.SubMenuItemID != nil {
			if err := s.detectCycle(*item.SubMenuItemID, ancestors); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *bomService) DeleteBOMItem(id uint) error {
	return s.bomRepo.DeleteItem(id)
}
