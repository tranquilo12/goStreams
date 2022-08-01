echo "Executing backfill from 2015-01-01 -> 2016-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2015-01-01 --to=2016-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2016-01-01 -> 2017-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2016-01-01 --to=2017-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2017-01-01 -> 2018-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2017-01-01 --to=2018-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2018-01-01 -> 2019-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2018-01-01 --to=2019-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2019-01-01 -> 2020-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2019-01-01 --to=2020-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2020-01-01 -> 2021-01-01"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2020-01-01 --to=2021-01-01 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"

echo "Executing backfill from 2021-01-01 -> 2021-06-15"
lightning aggsPub --dbtype=ec2db --timespan=day --mult=1 --from=2021-01-01 --to=2021-06-15 --forceInsertDate=2021-06-15 --limit=300 --useRedis=0;
echo "Done...\n"
echo "************************************************"
