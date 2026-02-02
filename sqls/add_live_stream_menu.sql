-- 添加直播流管理菜单项
-- 假设直播管理的父菜单ID为某个值，这里需要根据实际情况调整

-- 首先查找直播域名管理的菜单项，以便找到父菜单
-- SELECT * FROM sys_subsystem WHERE MENU_NAME LIKE '%直播%' OR DISPLAY_NAME LIKE '%直播%';

-- 添加直播流管理子系统/菜单项
-- 注意：需要根据实际的父菜单ID和公司ID进行调整
INSERT INTO sys_subsystem (
    ID,
    SYS_COMPANY_ID,
    MENU_NAME,
    DISPLAY_NAME,
    PARENT_ID,
    URL,
    ICON,
    ORDERNO,
    IS_ACTIVE,
    CREATE_BY,
    CREATE_TIME,
    UPDATE_BY,
    UPDATE_TIME
) VALUES (
    UUID(),                          -- ID
    (SELECT ID FROM sys_company LIMIT 1),  -- SYS_COMPANY_ID，使用第一个公司
    'LiveStream',                    -- MENU_NAME
    '直播流管理',                     -- DISPLAY_NAME
    (SELECT ID FROM sys_subsystem WHERE MENU_NAME = 'LiveDomain' LIMIT 1),  -- PARENT_ID，使用直播域名管理作为父菜单
    '/live/streams',                 -- URL
    'icon-play-circle',              -- ICON
    20,                              -- ORDERNO
    'Y',                             -- IS_ACTIVE
    'system',                        -- CREATE_BY
    NOW(),                           -- CREATE_TIME
    'system',                        -- UPDATE_BY
    NOW()                            -- UPDATE_TIME
);

-- 如果直播域名管理菜单不存在，可以先创建父菜单
-- INSERT INTO sys_subsystem (
--     ID,
--     SYS_COMPANY_ID,
--     MENU_NAME,
--     DISPLAY_NAME,
--     PARENT_ID,
--     URL,
--     ICON,
--     ORDERNO,
--     IS_ACTIVE,
--     CREATE_BY,
--     CREATE_TIME,
--     UPDATE_BY,
--     UPDATE_TIME
-- ) VALUES (
--     UUID(),
--     (SELECT ID FROM sys_company LIMIT 1),
--     'Live',
--     '直播管理',
--     NULL,
--     NULL,
--     'icon-live-broadcast',
--     100,
--     'Y',
--     'system',
--     NOW(),
--     'system',
--     NOW()
-- );
