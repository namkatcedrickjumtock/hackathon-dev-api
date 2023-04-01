<!-- Models -->
cars: {
    id:string
    name:string,
    engine_type:string
    car_model:string
    fuel_type:string,
    mileage:string,
    description:string.
    date_posted:string,
    seller_Id:string,
    catergory:string,
    photo url:string,
    biding_price:string,
    bid_expiration_time:string
}

bid:{
    id:string,
    card_id:string,
    bid_amount:string,
    user_name:string,
    user_email:string
}

user: {
    user_id:string,
    user_name:string,
    email:string
}


<!-- endpoints -->
GET  `/cars?start_key=0&count=10`{
    {
    id:string
    name:string,
    engine_type:string
    car_model:string
    fuel_type:string,
    mileage:string,
    description:string.
    date_posted:string,
    seller_Id:string,
    catergory:string,
    photo url:string,
    car_status:string
    biding_price:string,
    bid_expiration:string
}
}

GET  `/cars/:id`{}
POST  `/register/car`{

    <!-- req -->
     {
    id:string
    name:string,
    engine_type:string
    car_model:string
    fuel_type:string,
    mileage:string,
    description:string.
    date_posted:string,
    seller_Id:string,
    catergory:string,
    photo url:string,
    biding_price:string,
    bid_expiration:string
}
}
POST  `/bid/place_bid/:id`{
    card_id:string,
    amount:string,
    user_name:string,
    user_email:string
}

GET  `/user/:id`{}

PATCH  `/user/:id`{}


POST  `/user`{
    "user_id:string,
    user_email:string
}

POST `/webhook/campay/payments`{}   