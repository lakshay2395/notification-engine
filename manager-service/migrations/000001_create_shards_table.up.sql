CREATE TABLE IF NOT EXISTS shards
(
    id         UUID      NOT NULL,
    bucket_id  BIGINT    NOT NULL,
    created_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id, bucket_id),
    UNIQUE (id)
)