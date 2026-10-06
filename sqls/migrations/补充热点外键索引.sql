-- 为热点外键列补充普通索引（只加索引、不改数据，可随时 DROP INDEX 回退）
-- 覆盖：元数据加载与权限判断、数据字典、通用列表的公司隔离条件、主要业务主从关联
-- 已存在同名索引时该条会报 Duplicate key name，可忽略继续执行

CREATE INDEX idx_sys_column_sys_table_id ON sys_column (SYS_TABLE_ID);
CREATE INDEX idx_sys_user_groups_sys_user_id ON sys_user_groups (SYS_USER_ID);
CREATE INDEX idx_sys_group_prem_sys_groups_id ON sys_group_prem (SYS_GROUPS_ID);
CREATE INDEX idx_sys_group_prem_sys_directory_id ON sys_group_prem (SYS_DIRECTORY_ID);
CREATE INDEX idx_sys_directory_sys_table_id ON sys_directory (SYS_TABLE_ID);
CREATE INDEX idx_sys_dict_item_sys_dict_id ON sys_dict_item (SYS_DICT_ID);
CREATE INDEX idx_sys_table_sys_tablecategory_id ON sys_table (SYS_TABLECATEGORY_ID);
CREATE INDEX idx_sys_table_ref_sys_table_id ON sys_table_ref (SYS_TABLE_ID);
CREATE INDEX idx_sys_action_sys_table_id ON sys_action (SYS_TABLE_ID);
CREATE INDEX idx_sys_table_cmd_sys_table_id ON sys_table_cmd (SYS_TABLE_ID);
CREATE INDEX idx_sys_user_sys_company_id ON sys_user (SYS_COMPANY_ID);
CREATE INDEX idx_sys_groups_sys_company_id ON sys_groups (SYS_COMPANY_ID);
CREATE INDEX idx_sys_dict_sys_company_id ON sys_dict (SYS_COMPANY_ID);
CREATE INDEX idx_pull_stream_task_sys_company_id ON pull_stream_task (SYS_COMPANY_ID);
CREATE INDEX idx_cloud_upload_session_sys_company_id ON cloud_upload_session (SYS_COMPANY_ID);
CREATE INDEX idx_student_sys_company_id ON student (SYS_COMPANY_ID);
CREATE INDEX idx_teacher_sys_company_id ON teacher (SYS_COMPANY_ID);
CREATE INDEX idx_course_sys_company_id ON course (SYS_COMPANY_ID);
CREATE INDEX idx_score_sys_company_id ON score (SYS_COMPANY_ID);
CREATE INDEX idx_stu_class_sys_company_id ON stu_class (SYS_COMPANY_ID);
CREATE INDEX idx_cloud_upload_session_user_id ON cloud_upload_session (USER_ID);
CREATE INDEX idx_live_stream_room_id ON live_stream (ROOM_ID);
CREATE INDEX idx_pull_stream_task_room_id ON pull_stream_task (ROOM_ID);
CREATE INDEX idx_live_speech_task_room_id ON live_speech_task (ROOM_ID);
CREATE INDEX idx_score_student_id ON score (STUDENT_ID);
CREATE INDEX idx_score_course_id ON score (COURSE_ID);
CREATE INDEX idx_student_class_id ON student (CLASS_ID);
CREATE INDEX idx_sys_user_message_message_id ON sys_user_message (MESSAGE_ID);
CREATE INDEX idx_wf_instance_sys_table_id ON wf_instance (SYS_TABLE_ID);
