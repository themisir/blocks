ALTER TABLE posts
    ADD parent_id INTEGER NULL REFERENCES posts (id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE posts
    ADD children_count INTEGER NOT NULL DEFAULT 0;