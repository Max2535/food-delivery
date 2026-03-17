package service

import (
    "kitchen-service/internal/model"
    "kitchen-service/internal/repository"

    "github.com/rs/zerolog/log"
)

type KitchenService interface {
    CreateTicket(ticket *model.KitchenTicket) error
    UpdateStatus(orderID uint, status string) error
    GetByOrderID(orderID uint) (*model.KitchenTicket, error)
}

type kitchenService struct {
    repo repository.KitchenRepository
    // เพิ่ม RabbitMQ Producer เข้ามาตรงนี้ในอนาคต
}

func NewKitchenService(repo repository.KitchenRepository) KitchenService {
    return &kitchenService{repo: repo}
}

func (s *kitchenService) CreateTicket(ticket *model.KitchenTicket) error {
    ticket.Status = "Received"
    return s.repo.Create(ticket)
}

func (s *kitchenService) UpdateStatus(orderID uint, status string) error {
    err := s.repo.UpdateStatus(orderID, status)
    if err == nil && status == "Ready" {
        // เมื่อทำเสร็จ ให้เรียกฟังก์ชัน Publish Event ไปยัง RabbitMQ
        log.Info().Str("service", "kitchen-service").Uint("order_id", orderID).Msg("Order is Ready! Publishing event...")
        // PublishOrderReadyEvent(orderID) <- ฟังก์ชันที่เขียนค้างไว้คราวก่อน
    }
    return err
}

func (s *kitchenService) GetByOrderID(orderID uint) (*model.KitchenTicket, error) {
    return s.repo.GetByOrderID(orderID)
}
