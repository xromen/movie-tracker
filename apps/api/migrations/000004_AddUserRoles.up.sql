CREATE TABLE roles
(
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL UNIQUE
);

INSERT INTO roles(name) VALUES ('admin');

CREATE TABLE user_roles
(
    user_id    BIGINT       NOT NULL REFERENCES users (id),
    role_id    BIGINT       NOT NULL REFERENCES roles (id),
    PRIMARY KEY (user_id, role_id)
);