-- Add slug columns to content entities
ALTER TABLE articles ADD COLUMN IF NOT EXISTS slug VARCHAR(300);
ALTER TABLE article_categories ADD COLUMN IF NOT EXISTS slug VARCHAR(300);
ALTER TABLE forums ADD COLUMN IF NOT EXISTS slug VARCHAR(300);
ALTER TABLE forum_categories ADD COLUMN IF NOT EXISTS slug VARCHAR(300);
ALTER TABLE songs ADD COLUMN IF NOT EXISTS slug VARCHAR(300);
ALTER TABLE song_categories ADD COLUMN IF NOT EXISTS slug VARCHAR(300);

-- Add UUID columns to private entities
ALTER TABLE journals ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();
ALTER TABLE chat_sessions ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();
ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();
ALTER TABLE chat_folders ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();
ALTER TABLE playlists ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();
ALTER TABLE playlist_items ADD COLUMN IF NOT EXISTS uuid UUID DEFAULT gen_random_uuid();

-- Add username to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(50);

-- Populate slugs from titles/names for existing data
UPDATE articles SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(title, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;
UPDATE article_categories SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(name, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;
UPDATE forums SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(title, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;
UPDATE forum_categories SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(name, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;
UPDATE songs SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(title, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;
UPDATE song_categories SET slug = LOWER(REGEXP_REPLACE(REGEXP_REPLACE(name, '[^a-zA-Z0-9\s-]', '', 'g'), '\s+', '-', 'g')) || '-' || id WHERE slug IS NULL;

-- Populate UUIDs for existing data (make sure all rows have values)
UPDATE journals SET uuid = gen_random_uuid() WHERE uuid IS NULL;
UPDATE chat_sessions SET uuid = gen_random_uuid() WHERE uuid IS NULL;
UPDATE chat_messages SET uuid = gen_random_uuid() WHERE uuid IS NULL;
UPDATE chat_folders SET uuid = gen_random_uuid() WHERE uuid IS NULL;
UPDATE playlists SET uuid = gen_random_uuid() WHERE uuid IS NULL;
UPDATE playlist_items SET uuid = gen_random_uuid() WHERE uuid IS NULL;

-- Populate usernames from name + id
UPDATE users SET username = LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]', '', 'g')) || id WHERE username IS NULL;

-- Now add NOT NULL constraints and unique indexes
ALTER TABLE articles ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug) WHERE deleted_at IS NULL;

ALTER TABLE article_categories ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_article_categories_slug ON article_categories(slug) WHERE deleted_at IS NULL;

ALTER TABLE forums ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_forums_slug ON forums(slug) WHERE deleted_at IS NULL;

ALTER TABLE forum_categories ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_forum_categories_slug ON forum_categories(slug) WHERE deleted_at IS NULL;

ALTER TABLE songs ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_songs_slug ON songs(slug) WHERE deleted_at IS NULL;

ALTER TABLE song_categories ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_song_categories_slug ON song_categories(slug) WHERE deleted_at IS NULL;

ALTER TABLE journals ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_journals_uuid ON journals(uuid);

ALTER TABLE chat_sessions ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_sessions_uuid ON chat_sessions(uuid) WHERE deleted_at IS NULL;

ALTER TABLE chat_messages ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_messages_uuid ON chat_messages(uuid);

ALTER TABLE chat_folders ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_folders_uuid ON chat_folders(uuid);

ALTER TABLE playlists ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_playlists_uuid ON playlists(uuid) WHERE deleted_at IS NULL;

ALTER TABLE playlist_items ALTER COLUMN uuid SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_playlist_items_uuid ON playlist_items(uuid) WHERE deleted_at IS NULL;

ALTER TABLE users ALTER COLUMN username SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
