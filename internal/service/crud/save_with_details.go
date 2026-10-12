package crud

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
	"github.com/sky-xhsoft/sky-server/internal/pkg/transaction"
	"github.com/sky-xhsoft/sky-server/internal/service/groups"
	"gorm.io/gorm"
)

// detailPlan 一张明细表的保存计划（表、字段、指向主表的关联字段都已按元数据校验）
type detailPlan struct {
	table    *entity.SysTable
	columns  []*entity.SysColumn
	refField string
	records  []map[string]interface{}
}

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// SaveWithDetails 同时保存主表和明细（一个事务）。
// 主表和每张明细表都走与单条新增 / 修改相同的流程：表权限、公司与数据范围、字段处理（所属公司由服务端决定）、钩子。
// 明细表必须是主表的子表，关联字段只接受元数据里指向主表的字段；修改的明细必须属于这张主单。
func (s *service) SaveWithDetails(ctx context.Context, req *SaveWithDetailsRequest, userID uint) (*SaveWithDetailsResponse, error) {
	mainTable, err := s.metadataService.GetTable(req.TableName)
	if err != nil {
		return nil, errors.Wrap(errors.ErrResourceNotFound, "表不存在", err)
	}
	mainColumns, err := s.metadataService.GetColumns(mainTable.ID)
	if err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "获取列信息失败", err)
	}

	creating := req.Mode == "create"
	switch {
	case creating:
		if err := s.requirePermission(ctx, userID, mainTable, groups.PermCreate, "无创建权限"); err != nil {
			return nil, err
		}
	case req.Mode == "update":
		if req.MainRecordID == 0 {
			return nil, errors.New(errors.ErrInvalidParam, "缺少主表记录ID")
		}
		if err := s.requirePermission(ctx, userID, mainTable, groups.PermUpdate, "无修改权限"); err != nil {
			return nil, err
		}
		if err := s.ensureWritable(ctx, mainTable, mainColumns, userID, req.MainRecordID); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New(errors.ErrInvalidParam, "mode 只能是 create 或 update")
	}

	// 明细表与关联字段先全部校验，失败时什么都不写
	plans := make([]detailPlan, 0, len(req.Details))
	for _, d := range req.Details {
		dt, err := s.metadataService.GetTable(d.TableName)
		if err != nil {
			return nil, errors.Wrap(errors.ErrResourceNotFound, "子表不存在: "+d.TableName, err)
		}
		dcols, err := s.metadataService.GetColumns(dt.ID)
		if err != nil {
			return nil, errors.Wrap(errors.ErrInternal, "获取子表列信息失败", err)
		}
		ref, err := s.resolveDetailRefField(mainTable, dt, dcols, d.RefField)
		if err != nil {
			return nil, err
		}
		plans = append(plans, detailPlan{table: dt, columns: dcols, refField: ref, records: d.Records})
	}

	mainData := req.MainRecord
	if mainData == nil {
		mainData = map[string]interface{}{}
	}

	resp := &SaveWithDetailsResponse{Details: make([]DetailTableResponse, len(plans))}
	err = transaction.RunInTransaction(s.db, func(tx *gorm.DB) error {
		// 校验与写入都在同一事务里，才能看到刚插入、尚未提交的主表记录
		txs := *s
		txs.db = tx
		perms := map[string]bool{}

		var mainID uint
		if creating {
			processed, err := txs.prepareCreate(ctx, mainTable, mainColumns, mainData, userID)
			if err != nil {
				return err
			}
			if err := txs.insertWithHooks(ctx, tx, mainTable, mainData, processed); err != nil {
				return err
			}
			mainID = processed[colID].(uint)
			resp.MainRecord = processed
		} else {
			mainID = req.MainRecordID
			if hasBusinessFields(mainData) {
				processed, fields, err := txs.prepareUpdate(mainTable, mainColumns, mainID, mainData, userID)
				if err != nil {
					return err
				}
				if err := txs.updateWithHooks(ctx, tx, mainTable, mainID, mainData, processed, fields); err != nil {
					return err
				}
			}
			var cur map[string]interface{}
			if err := tx.Table(mainTable.Name).Where("ID = ?", mainID).Take(&cur).Error; err != nil {
				return errors.Wrap(errors.ErrDatabase, "查询主表记录失败", err)
			}
			resp.MainRecord = cur
		}

		for i, p := range plans {
			records, err := txs.saveDetail(ctx, tx, p, mainID, creating, userID, perms)
			if err != nil {
				return err
			}
			resp.Details[i] = DetailTableResponse{TableName: p.table.Name, Records: records}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// saveDetail 保存一张明细表：修改模式下，请求里没有带上的旧明细会被删除（与原行为一致）
func (s *service) saveDetail(ctx context.Context, tx *gorm.DB, p detailPlan, mainID uint, creating bool, userID uint, perms map[string]bool) ([]map[string]interface{}, error) {
	need := func(perm int, msg string) error {
		key := fmt.Sprintf("%d-%d", p.table.ID, perm)
		if perms[key] {
			return nil
		}
		if err := s.requirePermission(ctx, userID, p.table, perm, msg+"（"+p.table.Name+"）"); err != nil {
			return err
		}
		perms[key] = true
		return nil
	}
	refCond := p.refField + " = ?"

	if !creating {
		incoming := map[uint]bool{}
		for _, rec := range p.records {
			if hasID(rec) {
				incoming[toUintVal(rec[colID])] = true
			}
		}
		var existing []uint
		if err := tx.Table(p.table.Name).Where(refCond, mainID).Pluck(colID, &existing).Error; err != nil {
			return nil, errors.Wrap(errors.ErrDatabase, "查询已有子表记录失败", err)
		}
		var drop []uint
		for _, eid := range existing {
			if !incoming[eid] {
				drop = append(drop, eid)
			}
		}
		if len(drop) > 0 {
			if err := need(groups.PermDelete, "无删除明细权限"); err != nil {
				return nil, err
			}
			if err := s.ensureWritable(ctx, p.table, p.columns, userID, drop...); err != nil {
				return nil, err
			}
			for _, id := range drop {
				data := map[string]interface{}{colID: id}
				if err := s.executeHooksInTx(ctx, tx, p.table.ID, "D", "begin", data); err != nil {
					return nil, errors.Wrap(errors.ErrInternal, "执行before钩子失败", err)
				}
				if err := tx.Table(p.table.Name).Where("ID = ?", id).Delete(nil).Error; err != nil {
					return nil, errors.Wrap(errors.ErrDatabase, "删除子表记录失败", err)
				}
				if err := s.executeHooksInTx(ctx, tx, p.table.ID, "D", "end", data); err != nil {
					return nil, errors.Wrap(errors.ErrInternal, "执行after钩子失败", err)
				}
			}
		}
	}

	saved := make([]map[string]interface{}, 0, len(p.records))
	for _, rec := range p.records {
		if rec == nil {
			rec = map[string]interface{}{}
		}
		rec[p.refField] = mainID
		if !creating && hasID(rec) {
			rid := toUintVal(rec[colID])
			// 明细必须属于这张主单
			var n int64
			if err := tx.Table(p.table.Name).Where("ID = ?", rid).Where(refCond, mainID).Count(&n).Error; err != nil {
				return nil, errors.Wrap(errors.ErrDatabase, "查询子表记录失败", err)
			}
			if n == 0 {
				return nil, errors.New(errors.ErrResourceNotFound, "明细记录不存在或不属于这张单据")
			}
			if err := need(groups.PermUpdate, "无修改明细权限"); err != nil {
				return nil, err
			}
			if err := s.ensureWritable(ctx, p.table, p.columns, userID, rid); err != nil {
				return nil, err
			}
			delete(rec, colID)
			processed, fields, err := s.prepareUpdate(p.table, p.columns, rid, rec, userID)
			if err != nil {
				return nil, err
			}
			if err := s.updateWithHooks(ctx, tx, p.table, rid, rec, processed, fields); err != nil {
				return nil, err
			}
			var updated map[string]interface{}
			if err := tx.Table(p.table.Name).Where("ID = ?", rid).Take(&updated).Error; err != nil {
				return nil, errors.Wrap(errors.ErrDatabase, "查询子表记录失败", err)
			}
			saved = append(saved, updated)
			continue
		}
		delete(rec, colID)
		if err := need(groups.PermCreate, "无新增明细权限"); err != nil {
			return nil, err
		}
		processed, err := s.prepareCreate(ctx, p.table, p.columns, rec, userID)
		if err != nil {
			return nil, err
		}
		if err := s.insertWithHooks(ctx, tx, p.table, rec, processed); err != nil {
			return nil, err
		}
		saved = append(saved, processed)
	}
	return saved, nil
}

// resolveDetailRefField 明细表指向主表的关联字段：
// 只接受明细表中外键指向主表的字段，或 sys_table_ref 里登记的主表→明细关联字段；不接受请求里的任意字符串（防 SQL 注入）
func (s *service) resolveDetailRefField(mainTable, detail *entity.SysTable, detailColumns []*entity.SysColumn, requested string) (string, error) {
	allowed := map[string]bool{}
	for _, col := range detailColumns {
		if col.RefTableID != nil && *col.RefTableID == mainTable.ID {
			allowed[col.DbName] = true
		}
	}
	if refs, err := s.metadataService.GetTableRefs(mainTable.ID); err == nil {
		for _, ref := range refs {
			if uint(ref.RefTableID) != detail.ID {
				continue
			}
			name := mainTable.Name + "_ID"
			for _, col := range detailColumns {
				if ref.RefColumnID > 0 && col.ID == uint(ref.RefColumnID) {
					name = col.DbName
				}
			}
			allowed[name] = true
		}
	}
	if len(allowed) == 0 {
		return "", errors.New(errors.ErrInvalidParam, fmt.Sprintf("%s 不是 %s 的明细表", detail.Name, mainTable.Name))
	}
	if requested == "" && len(allowed) == 1 {
		for name := range allowed {
			requested = name
		}
	}
	if !allowed[requested] || !identRe.MatchString(requested) {
		return "", errors.New(errors.ErrInvalidParam, fmt.Sprintf("明细表 %s 的关联字段不正确", detail.Name))
	}
	return requested, nil
}

// hasBusinessFields 请求里是否有 ID 以外的字段（只改明细时主表不必更新）
func hasBusinessFields(data map[string]interface{}) bool {
	for k := range data {
		if k != colID && k != "id" {
			return true
		}
	}
	return false
}
