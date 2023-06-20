\c todo todo

-- Credit: https://www.timescale.com/blog/generating-more-realistic-sample-time-series-data-with-postgresql-generate_series/
/*
 * Function to create a random numeric value between two numbers
 * 
 * NOTICE: We are using the type of 'numeric' in this function in order
 * to visually return values that look like integers (no decimals) and 
 * floats (with decimals). However, if inserted into a table, the assumption
 * is that the appropriate column type is used. The `numeric` type is often
 * not the correct or most efficient type for storing numbers in a table.
 */
CREATE OR REPLACE FUNCTION random_between(min_val numeric, max_val numeric, round_to int=0)
    RETURNS numeric AS
$$
DECLARE
    value NUMERIC = random()* (min_val - max_val) + max_val;
BEGIN
    IF round_to = 0 THEN
        RETURN floor(value);
    ELSE
        RETURN round(value,round_to);
    END IF;
END
$$ language 'plpgsql';

-- Credit: https://dba.stackexchange.com/a/240446
CREATE OR REPLACE FUNCTION random_choice(
    choices text[]
)
    RETURNS text AS $$
DECLARE
    size_ int;
BEGIN
    size_ = array_length(choices, 1);
    RETURN (choices)[floor(random()*size_)+1];
END
$$ LANGUAGE plpgsql;

INSERT INTO lists
SELECT id,
       random_choice(array['Chores', 'Golang-Syd Meetup', 'Holiday', 'Moving House']) AS description
  FROM generate_series(1, 5000) AS id
;

SELECT setval('lists_id_seq', (SELECT MAX(id) + 1 FROM lists));

INSERT INTO items (
       id,
       list_id,
       description,
       due,
       completed
)
SELECT id,
       1 + ((900 * exp(abs(cos(id))))::int) % 1000 AS list_id,
       random_choice(array['Washing', 'Groceries', 'Pack Suitcase', 'Prepare Presentation']) AS description,
       CASE
           WHEN random() > 0.1 THEN NULL
           ELSE date_trunc('hour', now()) + random_between(-90, 90) * INTERVAL '1 day'
       END AS due,
       CASE
           WHEN random() > 0.25 THEN NULL
           ELSE date_trunc('hour', now()) + random_between(-90, 90) * INTERVAL '1 day'
       END AS completed
  FROM generate_series(1, 50000) WITH ORDINALITY AS t(id, rownum);
;

SELECT setval('items_id_seq', (SELECT MAX(id) + 1 FROM items));
