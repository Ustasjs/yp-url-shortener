ALTER TABLE short_urls
-- тут явно нужен внешний ключ - ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
-- к сожаление тулза wipedb в тестах на CI удаляет таблицы просто всем скопом и зависимости между ними вызывают ошибки
-- вернуть, когда понадобится работа с юзером
    ADD COLUMN user_id UUID;

CREATE INDEX IF NOT EXISTS idx_short_urls_user_id ON short_urls (user_id);