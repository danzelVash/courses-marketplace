-- +goose Up
-- +goose StatementBegin
create table if not exists users
(
    id            bigserial             not null unique,
    last_name     text                  not null,
    first_name    text                  not null,
    email         text,
    phone_number  text,
    password_hash text,
    salt          int,
    vk            boolean               not null,
    vk_id         bigint,
    is_admin      boolean default false not null
);

insert into users (last_name, first_name, vk, is_admin)
values ('admin', 'admin', false, true);

create table if not exists course
(
    id          serial not null unique,
    title       text   not null,
    description text,
    price       int    not null
);

create table if not exists users_courses
(
    id        serial                                           not null unique,
    user_id   integer references users (id) on delete cascade  not null unique,
    course_id integer references course (id) on delete cascade not null unique
);

create table if not exists session
(
    id           serial                                          not null,
    user_id      integer references users (id) on delete cascade not null,
    session      text,
    expired_date date                                            not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists session;
drop table if exists users_courses;
drop table if exists users;
drop table if exists course;
-- +goose StatementEnd
