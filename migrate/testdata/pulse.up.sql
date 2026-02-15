CREATE TABLE pulse_hourly (
    user_id   CHAR(26) NOT NULL,
    hostname  TEXT NOT NULL,
    stamp     DATETIME NOT NULL,
    count     INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (user_id, hostname, stamp)
);

CREATE TABLE pulse_daily (
    user_id   CHAR(26) NOT NULL,
    hostname  TEXT NOT NULL,
    stamp     DATE NOT NULL,
    count     INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (user_id, hostname, stamp)
);

CREATE TABLE pulse_hosts (
    user_id CHAR(26) NOT NULL,
    hostname TEXT NOT NULL,
    created_at DATETIME NOT NULL,

    PRIMARY KEY (user_id, hostname)
);
