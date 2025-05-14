BEGIN;
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;

create table public.client_limits
(
    id                             text     constraint client_limits_pk primary key,
    capacity                       integer  not null,
    refill_interval_milliseconds   bigint   not null
);

COMMIT;
