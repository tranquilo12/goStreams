CREATE USER grafanareader WITH PASSWORD 'Complicated@Password@Here';
GRANT USAGE ON SCHEMA public TO grafanareader;
GRANT SELECT ON public.aggregates_bars TO grafanareader;
GRANT SELECT ON public.ticker_vxes TO grafanareader;
GRANT SELECT ON public.ticker_news TO grafanareader;
GRANT SELECT ON public.snapshot_all_tickers TO grafanareader;
GRANT SELECT ON public.snapshot_gainers_losers TO grafanareader;
GRANT SELECT ON public.snapshot_one_tickers TO grafanareader;
GRANT SELECT ON public.previous_closes TO grafanareader;
GRANT SELECT ON public.markets TO grafanareader;
GRANT SELECT ON public.daily_open_closes TO grafanareader;
GRANT SELECT ON public.locales TO grafanareader;

CREATE INDEX CONCURRENTLY aggregates_bars__uindex
ON aggregates_bars(ticker, o, h, l, c, v, t, request_id, multiplier, timespan);

CREATE INDEX CONCURRENTLY ticker_vxes_ticker_market_last_updated_utc__uindex
    ON ticker_vxes(ticker, market, last_updated_utc);
