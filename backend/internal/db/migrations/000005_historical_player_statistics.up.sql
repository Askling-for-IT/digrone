BEGIN;

CREATE TABLE public.historical_player_statistics (
    player_id UUID NOT NULL REFERENCES public.players (id),
    competition TEXT NOT NULL,
    goals INTEGER NOT NULL,
    assists INTEGER NOT NULL,
    penalty_minutes INTEGER NOT NULL,
    source_key TEXT NOT NULL UNIQUE,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (player_id, competition)
);

ALTER TABLE public.matches
    ADD COLUMN historical_source_key TEXT UNIQUE;

COMMIT;
