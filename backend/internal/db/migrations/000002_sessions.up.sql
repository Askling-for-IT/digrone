CREATE TABLE public.sessions (
 token_hash TEXT PRIMARY KEY,
 issuer TEXT NOT NULL,
 subject TEXT NOT NULL,
 username TEXT NOT NULL,
 email TEXT NOT NULL,
 is_admin BOOLEAN NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expires_at_idx ON public.sessions (expires_at);
