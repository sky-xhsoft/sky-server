-- 统一表排序规则为 utf8mb4_0900_ai_ci（其余 61 张表已是该规则）
-- 混用 utf8mb4_unicode_ci 时跨表比较字符串会报 Illegal mix of collations
-- 两种规则均不区分大小写，转换不会产生新的唯一键冲突；如需回退改回 utf8mb4_unicode_ci 即可

ALTER TABLE cloud_item CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE cloud_quota CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE cloud_share CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE cloud_storage_config CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
ALTER TABLE live_room CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
