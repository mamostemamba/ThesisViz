package service

import (
	"github.com/google/uuid"

	"github.com/thesisviz/go-api/internal/model"
	"github.com/thesisviz/go-api/internal/repo"
)

type ProjectService struct {
	repo *repo.ProjectRepo
}

func NewProjectService(r *repo.ProjectRepo) *ProjectService {
	return &ProjectService{repo: r}
}

func (s *ProjectService) Create(p *model.Project) error {
	return s.repo.Create(p)
}

func (s *ProjectService) GetByID(id uuid.UUID) (*model.Project, error) {
	return s.repo.GetByID(id)
}

type PaginatedProjects struct {
	Items    []model.Project `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func (s *ProjectService) List(page, pageSize int) (*PaginatedProjects, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.List(offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &PaginatedProjects{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProjectService) Update(p *model.Project) error {
	return s.repo.Update(p)
}

func (s *ProjectService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}
