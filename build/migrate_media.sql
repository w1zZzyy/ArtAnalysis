-- Миграция: создание таблицы expert_media для хранения множественных медиафайлов (фото/видео)
-- Связь: expert_media.id_artcenter -> art_experts.id_artcenter

CREATE TABLE IF NOT EXISTS expert_media (
    id_media SERIAL PRIMARY KEY,
    id_artcenter INT NOT NULL REFERENCES art_experts(id_artcenter) ON DELETE CASCADE,
    media_url TEXT NOT NULL,
    media_type VARCHAR(20) NOT NULL DEFAULT 'image', -- 'image' или 'video'
    created_at TIMESTAMP DEFAULT NOW()
);

-- Индекс для быстрого поиска медиа по эксперту
CREATE INDEX IF NOT EXISTS idx_expert_media_artcenter ON expert_media(id_artcenter);

-- Комментарии
COMMENT ON TABLE expert_media IS 'Медиафайлы (фото/видео) для услуг art_experts';
COMMENT ON COLUMN expert_media.id_media IS 'Уникальный идентификатор медиафайла';
COMMENT ON COLUMN expert_media.id_artcenter IS 'Внешний ключ на услугу (эксперта)';
COMMENT ON COLUMN expert_media.media_url IS 'URL файла в MinIO';
COMMENT ON COLUMN expert_media.media_type IS 'Тип медиа: image или video';
COMMENT ON COLUMN expert_media.created_at IS 'Дата и время загрузки';
