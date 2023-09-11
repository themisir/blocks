CREATE TABLE posts
(
    id                integer   not null primary key autoincrement,
    content_markdown  text      not null,
    content_html      text      not null,
    is_anonymous      boolean   not null,
    author            text      null,
    depth             integer   not null default 0,
    top_level_post_id integer   null,
    parent_post_id    integer   null,
    created_at        timestamp not null default CURRENT_TIMESTAMP,
    deleted_at        timestamp null,

    FOREIGN KEY (parent_post_id) REFERENCES posts (id)
        ON DELETE SET NULL
        ON UPDATE RESTRICT
);

