package crud

import "math"

// unlimitedGrade 管理员不受字段访问级别限制
const unlimitedGrade = math.MaxInt32

// userFieldGrade 用户的字段访问级别：用户自身与其有效权限组 SGRADE 的最大值；管理员不受限制。
// 字段的 SGRADE 大于该值时，用户不能查看也不能修改该字段（SGRADE 为空/0 的字段不受限）。
func (s *service) userFieldGrade(userID uint) int {
	var row struct {
		IsAdmin   string `gorm:"column:IS_ADMIN"`
		UserGrade int    `gorm:"column:USER_GRADE"`
		GroupMax  int    `gorm:"column:GROUP_MAX"`
	}
	err := s.db.Raw(`SELECT u.IS_ADMIN, IFNULL(u.SGRADE, 0) AS USER_GRADE,
			IFNULL((SELECT MAX(g.SGRADE) FROM sys_user_groups ug
				JOIN sys_groups g ON g.ID = ug.SYS_DIRECTORY_ID AND g.IS_ACTIVE = 'Y'
				WHERE ug.SYS_USER_ID = u.ID AND ug.IS_ACTIVE = 'Y'), 0) AS GROUP_MAX
		FROM sys_user u WHERE u.ID = ?`, userID).Scan(&row).Error
	if err != nil {
		return 0
	}
	if row.IsAdmin == "Y" {
		return unlimitedGrade
	}
	if row.GroupMax > row.UserGrade {
		return row.GroupMax
	}
	return row.UserGrade
}

// fieldAllowed 字段访问级别是否允许
func fieldAllowed(colGrade, userGrade int) bool {
	return colGrade <= 0 || colGrade <= userGrade
}
