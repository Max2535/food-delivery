package service

import (
	"catalog-service/internal/model"
	"catalog-service/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
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
	repo  repository.MenuRepository
	redis *redis.Client
}

const (
	menuCacheKey    = "catalog:menus:all"
	menuItemKeyPrefix = "catalog:menu:"
	cacheTTL        = 10 * time.Minute
)

func NewMenuService(repo repository.MenuRepository, redisClient *redis.Client) MenuService {
	return &menuService{
		repo:  repo,
		redis: redisClient,
	}
}

func (s *menuService) GetAllMenuItems() ([]model.MenuItem, error) {
	ctx := context.Background()

	// Try to get from cache
	val, err := s.redis.Get(ctx, menuCacheKey).Result()
	if err == nil {
		var items []model.MenuItem
		if err := json.Unmarshal([]byte(val), &items); err == nil {
			log.Info().Msg("Cache hit: GetAllMenuItems")
			return items, nil
		}
	}

	// Cache miss, get from repo
	items, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// Save to cache
	data, _ := json.Marshal(items)
	s.redis.Set(ctx, menuCacheKey, data, cacheTTL)

	return items, nil
}

func (s *menuService) GetMenuItemByID(id uint) (*model.MenuItem, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("%s%d", menuItemKeyPrefix, id)

	// Try to get from cache
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var item model.MenuItem
		if err := json.Unmarshal([]byte(val), &item); err == nil {
			log.Info().Uint("id", id).Msg("Cache hit: GetMenuItemByID")
			return &item, nil
		}
	}

	// Cache miss
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Save to cache
	data, _ := json.Marshal(item)
	s.redis.Set(ctx, cacheKey, data, cacheTTL)

	return item, nil
}

func (s *menuService) CreateMenuItem(item *model.MenuItem) error {
	existing, _ := s.repo.FindByName(item.Name)
	if existing != nil {
		return ErrMenuItemAlreadyExists
	}
	if err := s.repo.Create(item); err != nil {
		return err
	}

	// Invalidate cache
	s.redis.Del(context.Background(), menuCacheKey)
	return nil
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

	// Invalidate cache
	ctx := context.Background()
	s.redis.Del(ctx, menuCacheKey)
	s.redis.Del(ctx, fmt.Sprintf("%s%d", menuItemKeyPrefix, id))

	return existing, nil
}

func (s *menuService) DeleteMenuItem(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Invalidate cache
	ctx := context.Background()
	s.redis.Del(ctx, menuCacheKey)
	s.redis.Del(ctx, fmt.Sprintf("%s%d", menuItemKeyPrefix, id))

	return nil
}
