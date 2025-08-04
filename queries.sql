-- queries.sql - SQL queries for sqlc code generation

-- get all migrations
SELECT filename FROM migrations ORDER BY filename;

-- record a migration
INSERT INTO migrations (filename) VALUES (?);