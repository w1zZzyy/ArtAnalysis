-- Миграция для добавления полей асинхронного анализа
-- Выполните этот скрипт после обновления кода

-- Добавляем новые поля для результата асинхронного анализа
ALTER TABLE center_requests 
ADD COLUMN IF NOT EXISTS analysis_result TEXT,
ADD COLUMN IF NOT EXISTS confidence_score REAL,
ADD COLUMN IF NOT EXISTS analysis_success BOOLEAN;

-- Создаём индекс для быстрого поиска заявок с результатами анализа
CREATE INDEX IF NOT EXISTS idx_center_requests_analysis ON center_requests(analysis_success) 
WHERE analysis_success IS NOT NULL;

COMMENT ON COLUMN center_requests.analysis_result IS 'Текстовый результат асинхронного анализа';
COMMENT ON COLUMN center_requests.confidence_score IS 'Показатель уверенности анализа (0.0-1.0)';
COMMENT ON COLUMN center_requests.analysis_success IS 'Успешность асинхронного анализа';
