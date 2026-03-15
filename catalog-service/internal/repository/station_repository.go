package repository

import (
	"catalog-service/internal/model"

	"gorm.io/gorm"
)

type StationRepository interface {
	FindAll() ([]model.KitchenStation, error)
	FindByID(id uint) (*model.KitchenStation, error)
	Create(station *model.KitchenStation) error
	AssignMenuToStation(mapping *model.MenuStationMapping) error
	FindStationsByMenuItemID(menuItemID uint) ([]model.KitchenStation, error)
}

type stationRepository struct {
	db *gorm.DB
}

func NewStationRepository(db *gorm.DB) StationRepository {
	return &stationRepository{db: db}
}

func (r *stationRepository) FindAll() ([]model.KitchenStation, error) {
	var stations []model.KitchenStation
	err := r.db.Find(&stations).Error
	return stations, err
}

func (r *stationRepository) FindByID(id uint) (*model.KitchenStation, error) {
	var station model.KitchenStation
	err := r.db.First(&station, id).Error
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *stationRepository) Create(station *model.KitchenStation) error {
	return r.db.Create(station).Error
}

func (r *stationRepository) AssignMenuToStation(mapping *model.MenuStationMapping) error {
	return r.db.Save(mapping).Error
}

func (r *stationRepository) FindStationsByMenuItemID(menuItemID uint) ([]model.KitchenStation, error) {
	var stations []model.KitchenStation
	err := r.db.
		Joins("JOIN menu_station_mappings ON menu_station_mappings.kitchen_station_id = kitchen_stations.id").
		Where("menu_station_mappings.menu_item_id = ?", menuItemID).
		Find(&stations).Error
	return stations, err
}
