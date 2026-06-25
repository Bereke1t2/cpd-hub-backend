-- Reversing the backfill means nulling out the email column for rows
-- that were backfilled (i.e. email == username and username looks like an email).
UPDATE users
   SET email = NULL
 WHERE email = username
   AND username LIKE '%@%';
