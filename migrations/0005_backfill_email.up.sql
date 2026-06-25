-- Phase 3 — backfill email column from username for rows created before Phase 2.
-- Rows where users.username holds an email (the old behaviour) are copied to the
-- email column so the new login query (WHERE email=$1 OR username=$1) works.
-- Safe to run multiple times; the WHERE email IS NULL guard is idempotent.

UPDATE users
   SET email = username
 WHERE email IS NULL
   AND username LIKE '%@%';   -- only rows that look like emails get backfilled
