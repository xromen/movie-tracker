CREATE TABLE reports
(
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 BIGINT NOT NULL REFERENCES users (id),
    period_from             DATE NOT NULL,
    period_to               DATE NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_schedule_create_at TIMESTAMPTZ NOT NULL,

    CHECK (period_from < period_to),
    UNIQUE (user_id, period_from, period_to)
);

CREATE TABLE report_messages
(
    id              BIGSERIAL PRIMARY KEY,
    report_id       BIGINT NOT NULL REFERENCES reports (id),
    message         TEXT NOT NULL,
    position        INTEGER NOT NULL,
    attempts        INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at         TIMESTAMPTZ,
    last_error      TEXT,

    UNIQUE (report_id, position)
);

CREATE INDEX idx_reports_next_schedule
    ON reports (user_id, next_schedule_create_at DESC);

CREATE INDEX idx_report_messages_pending
    ON report_messages (next_attempt_at)
    WHERE sent_at IS NULL;