package mysql

import (
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"gorm.io/gorm"
)

// sequenceRepository 序号生成器仓储MySQL实现
type sequenceRepository struct {
	db *gorm.DB
}

// NewSequenceRepository 创建序号生成器仓储
func NewSequenceRepository(db *gorm.DB) repository.SequenceRepository {
	return &sequenceRepository{db: db}
}

func (r *sequenceRepository) GetSequenceByName(name string) (*entity.SysSeq, error) {
	var seq entity.SysSeq
	err := r.db.Where("NAME = ? AND IS_ACTIVE = ?", name, "Y").First(&seq).Error
	if err != nil {
		return nil, err
	}
	return &seq, nil
}

func (r *sequenceRepository) UpdateSequence(seq *entity.SysSeq) error {
	return r.db.Save(seq).Error
}
