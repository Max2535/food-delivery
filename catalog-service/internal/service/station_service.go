package service

import (
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
)

type StationService interface {
	GetAllStations() ([]model.KitchenStation, error)
	CreateStation(station *model.KitchenStation) error
	AssignMenuToStation(menuItemID uint, stationID uint) error
	GetStationsByMenuItemID(menuItemID uint) ([]model.KitchenStation, error)
}

type stationService struct {
	repo     repository.StationRepository
	menuRepo repository.MenuRepository
}

func NewStationService(repo repository.StationRepository, menuRepo repository.MenuRepository) StationService {
	return &stationService{repo: repo, menuRepo: menuRepo}
}

func (s *stationService) GetAllStations() ([]model.KitchenStation, error) {
	return s.repo.FindAll()
}

func (s *stationService) CreateStation(station *model.KitchenStation) error {
	return s.repo.Create(station)
}

func (s *stationService) AssignMenuToStation(menuItemID uint, stationID uint) error {
	if _, err := s.menuRepo.FindByID(menuItemID); err != nil {
		return err
	}
	if _, err := s.repo.FindByID(stationID); err != nil {
		return err
	}
	return s.repo.AssignMenuToStation(&model.MenuStationMapping{
		MenuItemID:       menuItemID,
		KitchenStationID: stationID,
	})
}

func (s *stationService) GetStationsByMenuItemID(menuItemID uint) ([]model.KitchenStation, error) {
	return s.repo.FindStationsByMenuItemID(menuItemID)
}
