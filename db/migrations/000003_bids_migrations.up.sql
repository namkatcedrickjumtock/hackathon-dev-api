CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
    "bids" (
        "bid_id" uuid NOT NULL DEFAULT uuid_generate_v4 (),
        "car_id" VARCHAR(255),
        "bid_amount" VARCHAR(255) NOT NULL,
        "email" VARCHAR(255),
        "user_name" VARCHAR(255),
        "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY ("bid_id")
    );