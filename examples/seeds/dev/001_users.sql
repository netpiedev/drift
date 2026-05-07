INSERT INTO users(email) VALUES ('dev1@example.com') ON CONFLICT DO NOTHING;
INSERT INTO users(email) VALUES ('dev2@example.com') ON CONFLICT DO NOTHING;
