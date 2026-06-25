-- Phase 15 — Articles Feed. Copy to: migrations/0012_articles.up.sql

CREATE TABLE IF NOT EXISTS articles (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    author       TEXT NOT NULL DEFAULT '',
    source       TEXT NOT NULL DEFAULT 'CPD Hub',
    source_url   TEXT NOT NULL DEFAULT '',
    excerpt      TEXT NOT NULL DEFAULT '',
    full_content TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    rating       INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_articles_published ON articles(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source);

CREATE TABLE IF NOT EXISTS article_tags (
    article_id TEXT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (article_id, tag)
);
CREATE INDEX IF NOT EXISTS idx_article_tags_tag ON article_tags(tag);
