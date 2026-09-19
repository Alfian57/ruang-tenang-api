-- Remove all persisted data and schema owned by the retired Guild feature.
-- Keep the statements idempotent so this also works for databases created
-- before the Guild migrations were removed from the source tree.

DROP TABLE IF EXISTS guild_challenge_contributions CASCADE;
DROP TABLE IF EXISTS guild_activities CASCADE;
DROP TABLE IF EXISTS guild_challenges CASCADE;
DROP TABLE IF EXISTS guild_members CASCADE;
DROP TABLE IF EXISTS guilds CASCADE;
