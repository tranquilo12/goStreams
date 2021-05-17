-- May need to convert the date column to something else for index.
CREATE OR REPLACE FUNCTION datetime_to_date(dt timestamp with time zone)
    RETURNS date
AS
$BODY$
select CAST(dt AS date);
$BODY$
LANGUAGE sql
    IMMUTABLE;


-- Create index
drop index aggregates_bars_t_vw_multiplier_timespan_ticker_request_id_uind;

create unique index aggregates_bars_t_vw_multiplier_timespan_ticker_uind
    on aggregates_bars (t, vw, multiplier, timespan, ticker, o, h, l, c);
