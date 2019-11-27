CREATE DATABASE IF NOT EXISTS users;
CREATE TABLE IF NOT EXISTS users.users (
    created Date,
    login String,
    password String,
    role Enum('servant' = 1, 'master' = 2)
) ENGINE=MergeTree(created, (login), 8192);

CREATE TABLE IF NOT EXISTS users.tokens (
    created Date,
    login String,
    value String,
    expired DateTime
) ENGINE=MergeTree(created, (login), 8192);

CREATE DATABASE IF NOT EXISTS stats;
CREATE TABLE IF NOT EXISTS stats.stats (
    timestamp DateTime,
    type Enum('cpu' = 1, 'memory' = 2, 'la5' = 3),
    value Float64
) ENGINE=StripeLog();


