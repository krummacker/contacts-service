CREATE TABLE contacts (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    firstname   VARCHAR(50),
    lastname    VARCHAR(50),
    phone       VARCHAR(50),
    birthday    DATE
);

CREATE INDEX contacts_firstname
    ON contacts (firstname);

CREATE INDEX contacts_lastname
    ON contacts (lastname);