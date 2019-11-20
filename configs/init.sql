CREATE DATABASE IF NOT EXISTS users;
CREATE TABLE IF NOT EXISTS users.users (
    created Date,
    login String,
    password String
) ENGINE=MergeTree(created, (login), 8192);

CREATE TABLE IF NOT EXISTS users.tokens (
    login String,
    value String,
    expired DateTime
) ENGINE=TinyLog();




