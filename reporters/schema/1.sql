BEGIN;

ALTER TABLE listen
    ADD COLUMN geoip_country CHAR(2) NULL;

ALTER TABLE listen
    ADD COLUMN geoip_location TEXT NULL;

COMMIT;