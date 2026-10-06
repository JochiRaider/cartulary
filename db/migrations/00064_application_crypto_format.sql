-- +goose Up
CREATE TABLE public.application_crypto_format (
    singleton boolean NOT NULL,
    format_id text NOT NULL,
    CONSTRAINT application_crypto_format_pkey PRIMARY KEY (singleton),
    CONSTRAINT application_crypto_format_singleton_check CHECK (singleton)
);
REVOKE ALL ON TABLE public.application_crypto_format FROM cartulary_runtime, cartulary_recovery;
GRANT SELECT ON TABLE public.application_crypto_format TO cartulary_runtime, cartulary_recovery;

-- Identity is committed by admitted fresh initialization, never by a migration
-- applied to an existing deployment.

-- +goose Down
DROP TABLE public.application_crypto_format;
