-- Phase 3 placeholder.
-- ksqlDB streams and tables will be added when the stream processing
-- layer is built.
--
-- Planned:
--   CREATE STREAM user_events_raw (...) WITH (kafka_topic='cdc.user_events', ...);
--   CREATE STREAM user_logins AS SELECT ... WHERE event_type = 'LOGIN' EMIT CHANGES;
--   CREATE TABLE event_counts_5min AS ... WINDOW TUMBLING (SIZE 5 MINUTES) ...;

SET 'auto.offset.reset' = 'earliest';