USE test;

DROP TABLE contacts;

CREATE TABLE contacts (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    firstname   LONGTEXT,
    lastname    LONGTEXT,
    phone       LONGTEXT,
    birthday    DATE
);