package sequence

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/repository"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// Service 序号生成器服务接口
type Service interface {
	// 生成下一个编号
	NextValue(seqName string) (string, error)

	// 获取当前序号值（不递增）
	GetCurrentValue(seqName string) (string, error)

	// 重置序列
	ResetSequence(seqName string) error

	// 预览下一个编号（不实际生成）
	PreviewNext(seqName string) (string, error)
}

// service 序号生成器服务实现
type service struct {
	repo        repository.SequenceRepository
	redisClient *redis.Client
	ctx         context.Context
}

// NewService 创建序号生成器服务
func NewService(repo repository.SequenceRepository, redisClient *redis.Client) Service {
	return &service{
		repo:        repo,
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

// NextValue 生成下一个编号
func (s *service) NextValue(seqName string) (string, error) {
	// 使用分布式锁确保并发安全
	lockKey := fmt.Sprintf("lock:seq:%s", seqName)
	lock := s.redisClient.SetNX(s.ctx, lockKey, "1", 10*time.Second)

	if !lock.Val() {
		// 获取锁失败，等待重试
		time.Sleep(100 * time.Millisecond)
		return s.NextValue(seqName)
	}
	defer s.redisClient.Del(s.ctx, lockKey)

	// 查询序号生成器
	seq, err := s.repo.GetSequenceByName(seqName)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "序号生成器不存在", err)
	}

	// 获取当前日期
	now := time.Now()
	currentDate := s.getCurrentDateStr(seq.CycleType, now)

	// 检查是否需要重置
	if seq.CurDate != currentDate {
		seq.CurDate = currentDate
		seq.CurNum = 0
	}

	// 递增流水号
	seq.CurNum += seq.Incre

	// 格式化编号
	value := s.formatSequence(seq, now)

	// 更新数据库
	if err := s.repo.UpdateSequence(seq); err != nil {
		return "", errors.Wrap(errors.ErrDatabase, "更新序号失败", err)
	}

	return value, nil
}

// GetCurrentValue 获取当前序号值（不递增）
func (s *service) GetCurrentValue(seqName string) (string, error) {
	seq, err := s.repo.GetSequenceByName(seqName)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "序号生成器不存在", err)
	}

	// 获取当前日期
	now := time.Now()
	currentDate := s.getCurrentDateStr(seq.CycleType, now)

	// 如果周期已改变，返回新周期的初始值
	curNum := seq.CurNum
	if seq.CurDate != currentDate {
		curNum = 0
	}

	// 创建临时序号对象用于格式化
	tempSeq := &entity.SysSeq{
		VFormat: seq.VFormat,
		Prefix:  seq.Prefix,
		Suffix:  seq.Suffix,
		CurNum:  curNum,
	}

	return s.formatSequence(tempSeq, now), nil
}

// ResetSequence 重置序列
func (s *service) ResetSequence(seqName string) error {
	seq, err := s.repo.GetSequenceByName(seqName)
	if err != nil {
		return errors.Wrap(errors.ErrResourceNotFound, "序号生成器不存在", err)
	}

	seq.CurNum = 0
	seq.CurDate = ""

	return s.repo.UpdateSequence(seq)
}

// PreviewNext 预览下一个编号
func (s *service) PreviewNext(seqName string) (string, error) {
	seq, err := s.repo.GetSequenceByName(seqName)
	if err != nil {
		return "", errors.Wrap(errors.ErrResourceNotFound, "序号生成器不存在", err)
	}

	// 获取当前日期
	now := time.Now()
	currentDate := s.getCurrentDateStr(seq.CycleType, now)

	// 模拟递增
	nextNum := seq.CurNum
	if seq.CurDate != currentDate {
		nextNum = 0
	}
	nextNum += seq.Incre

	// 创建临时序号对象用于格式化
	tempSeq := &entity.SysSeq{
		VFormat: seq.VFormat,
		Prefix:  seq.Prefix,
		Suffix:  seq.Suffix,
		CurNum:  nextNum,
	}

	return s.formatSequence(tempSeq, now), nil
}

// getCurrentDateStr 获取当前日期字符串（根据循环类型）
func (s *service) getCurrentDateStr(cycleType string, now time.Time) string {
	switch cycleType {
	case "D": // 日
		return now.Format("20060102")
	case "M": // 月
		return now.Format("200601")
	case "Y": // 年
		return now.Format("2006")
	default: // N 不循环
		return ""
	}
}

// formatSequence 格式化序号
func (s *service) formatSequence(seq *entity.SysSeq, now time.Time) string {
	result := seq.VFormat

	// 替换前缀和后缀
	if seq.Prefix != "" {
		result = seq.Prefix + result
	}
	if seq.Suffix != "" {
		result = result + seq.Suffix
	}

	// 替换日期占位符
	result = strings.ReplaceAll(result, "{YYYY}", now.Format("2006"))
	result = strings.ReplaceAll(result, "{YY}", now.Format("06"))
	result = strings.ReplaceAll(result, "{MM}", now.Format("01"))
	result = strings.ReplaceAll(result, "{DD}", now.Format("02"))

	// 替换流水号占位符 {0000}
	re := regexp.MustCompile(`\{(0+)\}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// 提取0的个数
		zeros := strings.Trim(match, "{}")
		width := len(zeros)

		// 格式化流水号
		format := fmt.Sprintf("%%0%dd", width)
		return fmt.Sprintf(format, seq.CurNum)
	})

	return result
}
