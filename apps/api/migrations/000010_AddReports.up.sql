CREATE TABLE reports 
(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_schedule_create_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE report_messages
(
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES reports (id),
    message TEXT NOT NULL,
    order INT NOT NULL,
    UNIQUE (report_id, order)
);