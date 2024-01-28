CREATE TABLE embeddings (
    id SERIAL PRIMARY KEY,
    scope VARCHAR(50),
    combined TEXT,
    embeddings TEXT,
    n_tokens INT,
    created_at TIMESTAMP
);