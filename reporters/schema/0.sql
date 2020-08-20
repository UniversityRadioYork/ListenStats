BEGIN;

CREATE TABLE listen
(
    listen_id  bigserial not null,
    mount      text,
    client_id  text,
    ip_address inet,
    user_agent text,
    time_start timestamp with time zone,
    time_end   timestamp with time zone
);

COMMIT;
