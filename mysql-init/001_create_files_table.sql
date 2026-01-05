CREATE TABLE files (
                       id CHAR(36) NOT NULL PRIMARY KEY,
                       file_name VARCHAR(255) NOT NULL UNIQUE,
                       file_tags JSON NOT NULL,
                       file_size INT NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);