-- 更新 sys_column 的常用字段英文名称

-- 通用字段
UPDATE sys_column SET EName = 'ID' WHERE DISPLAY_NAME = 'ID';
UPDATE sys_column SET EName = 'Create Time' WHERE DISPLAY_NAME = '创建时间';
UPDATE sys_column SET EName = 'Update Time' WHERE DISPLAY_NAME = '更新时间';
UPDATE sys_column SET EName = 'Create By' WHERE DISPLAY_NAME = '创建人';
UPDATE sys_column SET EName = 'Update By' WHERE DISPLAY_NAME = '更新人';
UPDATE sys_column SET EName = 'Is Active' WHERE DISPLAY_NAME = '可用';
UPDATE sys_column SET EName = 'Company ID' WHERE DISPLAY_NAME = '所属公司';

-- 用户相关
UPDATE sys_column SET EName = 'Username' WHERE DISPLAY_NAME = '用户名';
UPDATE sys_column SET EName = 'True Name' WHERE DISPLAY_NAME = '真实姓名';
UPDATE sys_column SET EName = 'Phone' WHERE DISPLAY_NAME = '手机号';
UPDATE sys_column SET EName = 'Email' WHERE DISPLAY_NAME = '邮箱';
UPDATE sys_column SET EName = 'Password' WHERE DISPLAY_NAME = '密码';
UPDATE sys_column SET EName = 'Language' WHERE DISPLAY_NAME = '语言';
UPDATE sys_column SET EName = 'Is Admin' WHERE DISPLAY_NAME = '管理员';
UPDATE sys_column SET EName = 'Security Grade' WHERE DISPLAY_NAME = '信息安全等级';

-- 公司相关
UPDATE sys_column SET EName = 'Company Name' WHERE DISPLAY_NAME = '公司名称';
UPDATE sys_column SET EName = 'Company Code' WHERE DISPLAY_NAME = '公司编码';
UPDATE sys_column SET EName = 'Domain' WHERE DISPLAY_NAME = '公司域名';
UPDATE sys_column SET EName = 'Description' WHERE DISPLAY_NAME = '描述';

-- 元数据相关
UPDATE sys_column SET EName = 'Table Name' WHERE DISPLAY_NAME = '表名';
UPDATE sys_column SET EName = 'Display Name' WHERE DISPLAY_NAME = '显示名称';
UPDATE sys_column SET EName = 'Column Name' WHERE DISPLAY_NAME = '字段名';
UPDATE sys_column SET EName = 'Column Type' WHERE DISPLAY_NAME = '字段类型';
UPDATE sys_column SET EName = 'Column Length' WHERE DISPLAY_NAME = '字段长度';
UPDATE sys_column SET EName = 'Precision' WHERE DISPLAY_NAME = '字段精度';
UPDATE sys_column SET EName = 'Nullable' WHERE DISPLAY_NAME = '是否允许NULL';
UPDATE sys_column SET EName = 'Is Display Key' WHERE DISPLAY_NAME = '是否显示键';
UPDATE sys_column SET EName = 'Is Input Key' WHERE DISPLAY_NAME = '是否输入键';
UPDATE sys_column SET EName = 'Uppercase' WHERE DISPLAY_NAME = '是否大写';
UPDATE sys_column SET EName = 'Query' WHERE DISPLAY_NAME = '是否查询条件';
UPDATE sys_column SET EName = 'Order No' WHERE DISPLAY_NAME = '排序号';
UPDATE sys_column SET EName = 'Mask' WHERE DISPLAY_NAME = '字段权限掩码';
UPDATE sys_column SET EName = 'Editable' WHERE DISPLAY_NAME = '是否可修改';
UPDATE sys_column SET EName = 'Set Value Type' WHERE DISPLAY_NAME = '字段赋值类型';
UPDATE sys_column SET EName = 'Reference Table ID' WHERE DISPLAY_NAME = '引用表ID';
UPDATE sys_column SET EName = 'Reference Column ID' WHERE DISPLAY_NAME = '引用列ID';
UPDATE sys_column SET EName = 'Reference On Delete' WHERE DISPLAY_NAME = '级联删除配置';
UPDATE sys_column SET EName = 'Sequence Code' WHERE DISPLAY_NAME = '序号编码';
UPDATE sys_column SET EName = 'Data Dictionary ID' WHERE DISPLAY_NAME = '字典ID';
UPDATE sys_column SET EName = 'Default Value' WHERE DISPLAY_NAME = '默认值';
UPDATE sys_column SET EName = 'Regular Expression' WHERE DISPLAY_NAME = '正则表达式';
UPDATE sys_column SET EName = 'Error Message' WHERE DISPLAY_NAME = '错误信息';
UPDATE sys_column SET EName = 'Filter' WHERE DISPLAY_NAME = '过滤条件';
UPDATE sys_column SET EName = 'Display Type' WHERE DISPLAY_NAME = '显示类型';
UPDATE sys_column SET EName = 'Display Columns' WHERE DISPLAY_NAME = '显示列数';
UPDATE sys_column SET EName = 'Display Rows' WHERE DISPLAY_NAME = '显示行数';
UPDATE sys_column SET EName = 'Properties' WHERE DISPLAY_NAME = '扩展属性';
UPDATE sys_column SET EName = 'Show Title' WHERE DISPLAY_NAME = '是否显示标题';
UPDATE sys_column SET EName = 'Submethod' WHERE DISPLAY_NAME = '小计类型';
UPDATE sys_column SET EName = 'HR Column ID' WHERE DISPLAY_NAME = 'HR关联列ID';
UPDATE sys_column SET EName = 'Show Column ID' WHERE DISPLAY_NAME = '展示表ID';
UPDATE sys_column SET EName = 'Show Column Value' WHERE DISPLAY_NAME = '展示表字段值';
UPDATE sys_column SET EName = 'Column Grade' WHERE DISPLAY_NAME = '字段安全等级';

-- 直播相关
UPDATE sys_column SET EName = 'Live Domain' WHERE DISPLAY_NAME = '直播域名';
UPDATE sys_column SET EName = 'Live Stream' WHERE DISPLAY_NAME = '直播流';
UPDATE sys_column SET EName = 'Live Room' WHERE DISPLAY_NAME = '直播间';
UPDATE sys_column SET EName = 'Pull Stream Task' WHERE DISPLAY_NAME = '拉流任务';
UPDATE sys_column SET EName = 'Live Recording' WHERE DISPLAY_NAME = '直播录制';

-- 云盘相关
UPDATE sys_column SET EName = 'Cloud Item' WHERE DISPLAY_NAME = '云盘';
UPDATE sys_column SET EName = 'File Name' WHERE DISPLAY_NAME = '文件名';
UPDATE sys_column SET EName = 'File Size' WHERE DISPLAY_NAME = '文件大小';
UPDATE sys_column SET EName = 'File Type' WHERE DISPLAY_NAME = '文件类型';

-- 查询验证
SELECT ID, SYS_TABLE_ID, DISPLAY_NAME, EName
FROM sys_column
WHERE EName IS NOT NULL
ORDER BY ID
LIMIT 50;
