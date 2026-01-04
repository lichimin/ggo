-- 存档表添加area字段
ALTER TABLE archives ADD COLUMN IF NOT EXISTS area INTEGER NOT NULL DEFAULT 1;

-- 创建区服表
CREATE TABLE IF NOT EXISTS areas (
    id SERIAL PRIMARY KEY,
    area INTEGER NOT NULL UNIQUE,
    is_new BOOLEAN NOT NULL DEFAULT false,
    status INTEGER NOT NULL DEFAULT 1,
    name VARCHAR(50),
    max_users INTEGER NOT NULL DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_archives_area ON archives(area);
CREATE INDEX IF NOT EXISTS idx_areas_area ON areas(area);

-- 插入默认区服数据
INSERT INTO areas (area, is_new, status, name, max_users) VALUES 
(1, false, 1, '一区', 1000),
(2, true, 1, '二区', 1000),
(3, true, 1, '三区', 1000),
(4, false, 1, '四区', 1000),
(5, false, 1, '五区', 1000)
ON CONFLICT (area) DO NOTHING;