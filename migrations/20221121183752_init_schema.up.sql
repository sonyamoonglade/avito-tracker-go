CREATE TABLE IF NOT EXISTS "adverts"(
    "advert_id" UUID PRIMARY KEY UNIQUE,
    "url" varchar(255) UNIQUE NOT NULL,
    "title" varchar(255) NOT NULL DEFAULT '',
    "last_price" REAL NOT NULL DEFAULT 0.0,
    "current_price" REAL NOT NULL DEFAULT 0.0,
    "is_parsed" BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS "subscribers"(
    "subscriber_id" UUID PRIMARY KEY UNIQUE,
    "telegram_id" INTEGER NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "subscriptions"(
    "advert_id" UUID NOT NULL,
    "subscriber_id" UUID NOT NULL
);

ALTER TABLE "subscriptions" ADD CONSTRAINT "advert_id_fk"
    FOREIGN KEY("advert_id")
    REFERENCES adverts("advert_id")
    ON DELETE CASCADE;

ALTER TABLE "subscriptions" ADD CONSTRAINT "subscriber_id_fk"
    FOREIGN KEY("subscriber_id")
    REFERENCES subscribers("subscriber_id")
    ON DELETE CASCADE;



