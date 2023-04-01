CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  "cars" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4 (),
    "properties" jsonb NOT NULL,
    PRIMARY KEY ("id")
  );