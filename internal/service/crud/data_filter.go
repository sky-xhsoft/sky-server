package crud

import "context"

// userTableDataFilter 用户在表对应安全目录上的权限组数据过滤条件；表没有安全目录时不过滤
func (s *service) userTableDataFilter(ctx context.Context, userID, tableID uint) (map[string]interface{}, error) {
	var directoryIDs []uint
	if err := s.db.WithContext(ctx).Table("sys_directory").
		Where("SYS_TABLE_ID = ? AND IS_ACTIVE = ?", tableID, "Y").
		Order("ID").Limit(1).
		Pluck("ID", &directoryIDs).Error; err != nil {
		return nil, err
	}
	if len(directoryIDs) == 0 {
		return nil, nil
	}
	return s.groupsService.GetUserDataFilter(ctx, userID, directoryIDs[0])
}
