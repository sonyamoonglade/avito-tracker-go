CREATE TABLE IF NOT EXISTS "adverts"(
    "advert_id" SERIAL PRIMARY KEY,
    "url" varchar(255) UNIQUE NOT NULL,
    "last_price" REAL NOT NULL,
    "current_price" REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS "subscribers"(
    "subscriber_id" SERIAL PRIMARY KEY,
    "telegram_id" INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "subscriptions"(
    "advert_id" INTEGER NOT NULL,
    "subscriber_id" INTEGER NOT NULL
);

ALTER TABLE "subscriptions" ADD CONSTRAINT "advert_id_fk"
    FOREIGN KEY("advert_id")
    REFERENCES adverts("advert_id")
    ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD CONSTRAINT "subscriber_id_fk"
    FOREIGN KEY("subscriber_id")
    REFERENCES subscribers("subscriber_id")
    ON DELETE CASCADE;



