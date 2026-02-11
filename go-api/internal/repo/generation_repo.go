package repo

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/thesisviz/go-api/internal/model"
)

type GenerationRepo struct {
	db *gorm.DB
}

func NewGenerationRepo(db *gorm.DB) *GenerationRepo {
	return &GenerationRepo{db: db}
}

func (r *GenerationRepo) Create(g *model.Generation) error {
	return r.db.Create(g).Error
}

func (r *GenerationRepo) GetByID(id uuid.UUID) (*model.Generation, error) {
	var g model.Generation
	if err := r.db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GenerationRepo) ListByProject(projectID uuid.UUID, offset, limit int) ([]model.Generation, int64, error) {
	var generations []model.Generation
	var total int64

	q := r.db.Model(&model.Generation{}).Where("project_id = ?", projectID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&generations).Error; err != nil {
		return nil, 0, err
	}
	return generations, total, nil
}

func (r *GenerationRepo) Update(g *model.Generation) error {
	return r.db.Save(g).Error
}

func (r *GenerationRepo) Delete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&model.Generation{}, "id = ?", id).Error
}
