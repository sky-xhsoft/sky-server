package repository

import "github.com/sky-xhsoft/sky-server/internal/model/entity"

// SequenceRepository 序号生成器仓储接口
type SequenceRepository interface {
	// 根据名称获取序号生成器
	GetSequenceByName(name string) (*entity.SysSeq, error)

	// 更新序号生成器
	UpdateSequence(seq *entity.SysSeq) error
}
