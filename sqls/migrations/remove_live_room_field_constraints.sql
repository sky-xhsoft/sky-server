-- 移除直播间表字段的 NOT NULL 约束
-- 为前端删除的字段取消必填约束

USE `skyserver`;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 移除 ROOM_TYPE 的 NOT NULL 约束，添加默认值
ALTER TABLE `live_room`
MODIFY COLUMN `ROOM_TYPE` varchar(50) DEFAULT 'video' COMMENT '直播间类型：video(视频直播), image(图片直播), vr(VR直播), audio(语音直播), graphic(图文直播)',
MODIFY COLUMN `ROOM_STAGE` varchar(50) DEFAULT 'formal' COMMENT '直播间阶段：formal(正式直播), test(测试直播)',
MODIFY COLUMN `VIEWING_METHOD` varchar(50) DEFAULT 'public' COMMENT '观看方式：public(公开), encrypted(加密), paid(付费), ticket(购票进入), enterprise(企业成员观看), custom(自建成员观看)',
MODIFY COLUMN `PLAYBACK_METHOD` varchar(50) DEFAULT 'post_end' COMMENT '回放方式：post_end(结束后回放), real_time(实时回放), no_playback(结束后不回放)',
MODIFY COLUMN `PLAYBACK_VALIDITY` varchar(50) DEFAULT 'unlimited' COMMENT '回放有效期：unlimited(无限制), all_day(全天), partial(部分时段)';

SET FOREIGN_KEY_CHECKS = 1;
