INSERT INTO users(email) VALUES ('staging@example.com') ON CONFLICT DO NOTHING;
