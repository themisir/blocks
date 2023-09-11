CREATE TABLE IF NOT EXISTS posts
(
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    content_markdown TEXT      NOT NULL,
    content_html     TEXT      NOT NULL,
    author           TEXT      NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMP NULL
);