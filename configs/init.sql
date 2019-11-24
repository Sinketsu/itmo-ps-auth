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

CREATE DATABASE IF NOT EXISTS stats;
CREATE TABLE IF NOT EXISTS stats.stats (
    timestamp DateTime,
    type Enum('cpu' = 1, 'memory' = 2, 'la5' = 3),
    value Float64
) ENGINE=StripeLog();


