DROP TABLE IF EXISTS sinners;

CREATE TABLE sinners (
    code           BIGINT PRIMARY KEY,
    "name"         VARCHAR(20) NOT NULL,
    class          CHAR(1) NOT NULL,
    libram         VARCHAR(20),
    tendency       VARCHAR(20),
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON sinners
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

INSERT INTO sinners (code, "name", class, libram, tendency)
VALUES
    (14, 'Deren', 'S', 'Fraud', 'Fury'),
    (17, 'Shalom', 'S', 'Sloth', 'Reticle');

CREATE PUBLICATION dbz_publication FOR ALL TABLES;
