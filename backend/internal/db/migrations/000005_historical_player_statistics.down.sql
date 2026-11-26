BEGIN;

DROP TABLE public.historical_player_statistics;
ALTER TABLE public.matches DROP COLUMN historical_source_key;

COMMIT;
