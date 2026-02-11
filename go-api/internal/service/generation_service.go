package service

import (
	"github.com/google/uuid"

	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/repo"
)

type GenerationService struct {
	repo *repo.GenerationRepo
}

func NewGenerationService(r *repo.GenerationRepo) *GenerationService {
	return &GenerationService{repo: r}
}

func (s *GenerationService) Create(g *model.Generation) error {
	return s.repo.Create(g)
}

func (s *GenerationService) GetByID(id uuid.UUID) (*model.Generation, error) {
	return s.repo.GetByID(id)
}

type PaginatedGenerations struct {
	Items    []model.Generation `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

func (s *GenerationService) ListByProject(projectID uuid.UUID, page, pageSize int) (*PaginatedGenerations, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.ListByProject(projectID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &PaginatedGenerations{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GenerationService) Update(g *model.Generation) error {
	return s.repo.Update(g)
}

func (s *GenerationService) UpdateStatus(id uuid.UUID, status string) error {
	g, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	g.Status = status
	return s.repo.Update(g)
}

func (s *GenerationService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
