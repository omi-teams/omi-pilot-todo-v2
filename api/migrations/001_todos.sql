CREATE TABLE IF NOT EXISTS todos (
    id bigint GENERATED ALWAYS AS IDENTITY CONSTRAINT todos_pkey PRIMARY KEY,
    title text NOT NULL CONSTRAINT todos_title_check CHECK (
        char_length(title) BETWEEN 1 AND 200 AND title !~ '^[[:space:]]|[[:space:]]$'
    ),
    done boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS todos_created_at_id_idx ON todos (created_at DESC, id DESC);
