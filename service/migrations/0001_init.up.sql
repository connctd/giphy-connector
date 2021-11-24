CREATE TABLE installations (
    id CHAR (36) NOT NULL,
    token TEXT NOT NULL,
    UNIQUE(id)
);

CREATE TABLE instances (
    id CHAR (36) NOT NULL,
    token TEXT NOT NULL,
    installation_id CHAR (36) NOT NULL,
    thing_id CHAR (36) NOT NULL DEFAULT '',
    UNIQUE(id),
    FOREIGN KEY (installation_id)
        REFERENCES installations(id)
);