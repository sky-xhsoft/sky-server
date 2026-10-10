-- 补全分片上传遗留的空 PATH（旧版分片上传完成时未写路径，重命名会因此报 500）
-- 子目录中的文件：父目录 PATH + '/' + 名称；根目录文件：'/' + 名称
UPDATE cloud_item c
JOIN cloud_item p ON p.ID = c.PARENT_ID
SET c.PATH = CONCAT(TRIM(TRAILING '/' FROM p.PATH), '/', c.NAME), c.UPDATE_TIME = NOW()
WHERE c.ITEM_TYPE = 'file' AND (c.PATH IS NULL OR c.PATH = '') AND p.PATH IS NOT NULL AND p.PATH <> '';

UPDATE cloud_item
SET PATH = CONCAT('/', NAME), UPDATE_TIME = NOW()
WHERE ITEM_TYPE = 'file' AND (PATH IS NULL OR PATH = '') AND PARENT_ID IS NULL;
