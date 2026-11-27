BEGIN;

CREATE TABLE public.players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL
);

CREATE TABLE public.matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    start_time TIMESTAMPTZ NOT NULL,
    opponent TEXT NOT NULL,
    home_game BOOLEAN NOT NULL DEFAULT false,
    competition TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled'
        CHECK (status IN ('scheduled', 'completed', 'cancelled')),
    goals_for INTEGER,
    goals_against INTEGER
);

CREATE TABLE public.events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID REFERENCES public.matches (id),
    player_id UUID NOT NULL REFERENCES public.players (id),
    event_type TEXT NOT NULL
        CHECK (event_type IN ('goal', 'penalty', 'fine', 'fine_payment')),
    assisting_player_id UUID REFERENCES public.players (id),
    penalty_minutes INTEGER,
    num_beers INTEGER,
    comment TEXT,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX matches_competition_start_time_idx
    ON public.matches (competition, start_time);
CREATE INDEX events_match_id_idx ON public.events (match_id);
CREATE INDEX events_player_id_idx ON public.events (player_id);
CREATE INDEX events_assisting_player_id_idx ON public.events (assisting_player_id)
    WHERE assisting_player_id IS NOT NULL;

COMMIT;
