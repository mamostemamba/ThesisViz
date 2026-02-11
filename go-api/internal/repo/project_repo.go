package repo

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/model"
)

type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(p *model.Project) error {
	return r.db.Create(p).Error
}

func (r *ProjectRepo) GetByID(id uuid.UUID) (*model.Project, error) {
	var p model.Project
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) List(offset, limit int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	if err := r.db.Model(&model.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&projects).Error; err != nil {
		return nil, 0, err
	}
	return projects, total, nil
}

func (r *ProjectRepo) Update(p *model.Project) error {
	return r.db.Save(p).Error
}

func (r *ProjectRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Project{}, "id = ?", id).Error
}
