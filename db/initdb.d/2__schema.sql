\c todo todo

CREATE SCHEMA todo AUTHORIZATION todo;

CREATE TABLE lists(
  id          SERIAL PRIMARY KEY,
  description TEXT
);

CREATE TABLE items(
  id          SERIAL    PRIMARY KEY,
  list_id     INT       NOT NULL,
  description TEXT      NOT NULL,
  due         TIMESTAMP,
  completed   TIMESTAMP,
  FOREIGN KEY (list_id) REFERENCES lists (id)
);
