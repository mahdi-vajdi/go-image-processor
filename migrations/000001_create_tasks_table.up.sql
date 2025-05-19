CREATE TABLE image_processing_tasks
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    original_filename VARCHAR(255) NOT NULL,
    storage_key       VARCHAR(255) NOT NULL,
    status            VARCHAR(50)  NOT NULL,
    error_message     TEXT,
    created_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_status ON image_processing_tasks (status);
