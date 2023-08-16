CREATE TABLE users
(
    id                  integer   not null primary key autoincrement,
    username            text      not null,
    normalized_username text      not null unique,
    password_hash       text      not null,
    role                text      not null default 'user',
    created_at          timestamp not null default CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_username_idx ON users (normalized_username);

CREATE TABLE user_sessions
(
    session_id text      not null primary key,
    user_id    integer   not null,
    user_agent text      null,
    ip_address text      null,
    created_at timestamp not null default CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE SET NULL
);

CREATE TABLE posts
(
    id                integer   not null primary key autoincrement,
    content_markdown  text      not null,
    content_html      text      not null,
    is_anonymous      boolean   not null,
    user_id           integer   null,
    depth             integer   not null default 0,
    top_level_post_id integer   null,
    parent_post_id    integer   null,
    created_at        timestamp not null default CURRENT_TIMESTAMP,
    deleted_at        timestamp null,

    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE
        ON UPDATE RESTRICT,
    FOREIGN KEY (parent_post_id) REFERENCES posts (id)
        ON DELETE SET NULL
        ON UPDATE RESTRICT
);

