CREATE TABLE installation_configuration (
    installation_id CHAR (36) NOT NULL,
    id CHAR (36) NOT NULL,
    value VARCHAR (200) NOT NULL,
    FOREIGN KEY (installation_id)
        REFERENCES installations(id)
);