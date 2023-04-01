CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE
  "cars" (
    "id" uuid NOT NULL DEFAULT uuid_generate_v4 (),
    "properties" jsonb NOT NULL,
    PRIMARY KEY ("id")
  );

INSERT INTO
    cars (properties)
VALUES
    ('{
  "car_name": "BMW 7 Series",
  "city_id": "douala",
  "seller_id": "",
  "date_posted": "2022-11-28",
  "time": "04:50:00",
  "photo_url": "https://images.unsplash.com/photo-1627936354732-ffbe552799d8?ixlib=rb-4.0.3&ixid=MnwxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8&auto=format&fit=crop&w=924&q=80",
  "description": "The flagship BMW earned its reputation for stately elegance years ago. Now, the 2023 BMW 7 Series Sedans offer a new take on luxury automobile design and start a new chapter in their legacy",
  "category_id": "Sedan"
}'
    );