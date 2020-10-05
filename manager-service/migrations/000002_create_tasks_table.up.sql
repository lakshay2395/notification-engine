CREATE TABLE IF NOT EXISTS tasks
(
    id                 UUID         NOT NULL,
    shard_id           UUID         NOT NULL,
    status             VARCHAR(50)  NOT NULL,
    source             VARCHAR(200) NOT NULL,
    destination        VARCHAR(200) NOT NULL,
    time_for_execution TIMESTAMP    NOT NULL,
    content            text         NOT NULL,
    PRIMARY KEY (id, shard_id),
    FOREIGN KEY (shard_id) REFERENCES shards (id)
)