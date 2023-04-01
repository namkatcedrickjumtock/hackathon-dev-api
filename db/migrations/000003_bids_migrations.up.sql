CREATE TABLE
    "bids" (
        "bid_id" VARCHAR(255) NOT NULL,
        "car_id" VARCHAR(255),
        "bid_amount" VARCHAR(255) NOT NULL,
        "email" VARCHAR(255),
        "user_name" VARCHAR(255),
        PRIMARY KEY ("bid_id")
    );