CREATE TABLE IF NOT EXISTS processed_image_details
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id        BIGINT       NOT NULL,
    format         VARCHAR(10)  NOT NULL,
    size           VARCHAR(20)  NOT NULL,
    storage_key    VARCHAR(255) NOT NULL,
    result_message TEXT,
    created_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES image_processing_tasks (id)
);

CREATE INDEX IF NOT EXISTS idx_details_task_id ON processed_images (task_id)
