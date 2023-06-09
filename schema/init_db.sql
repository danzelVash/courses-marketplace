CREATE TABLE users
(
    id            bigserial             not null unique,
    last_name     varchar(250)          not null,
    first_name    varchar(250)          not null,
    email         varchar(250),
    phone_number  varchar(250),
    password_hash varchar(250),
    salt          int,
    vk            boolean               not null,
    vk_id         bigint,
    is_admin      boolean default false not null
);

INSERT INTO users (last_name, first_name, vk, is_admin)
VALUES ('admin', 'admin', false, true);

CREATE TABLE course
(
    id          serial       not null unique,
    title       varchar(250) not null,
    description text,
    price       int          not null
);

CREATE TABLE users_courses
(
    id        serial                                           not null unique,
    user_id   integer references users (id) on delete cascade  not null unique,
    course_id integer references course (id) on delete cascade not null unique
);

CREATE TABLE item
(
    id           serial       not null unique,
    path_to_item varchar(250) not null unique
);

CREATE TABLE items_courses
(
    id        serial                                           not null unique,
    course_id integer references course (id) on delete cascade not null unique,
    item_id   integer references item (id) on delete cascade   not null unique
);

CREATE TABLE session
(
    id           serial                                          not null,
    user_id      integer references users (id) on delete cascade not null,
    session      varchar(250),
    expired_date date                                            not null
);