ALTER TABLE public.events ADD CONSTRAINT events_match_link_check CHECK (
    (event_type IN ('goal', 'penalty') AND match_id IS NOT NULL)
    OR (event_type IN ('fine', 'fine_payment') AND match_id IS NULL)
);
