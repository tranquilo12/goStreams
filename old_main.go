package main

//func ParseTime(date string) string {
//	// Convert all dates to a PROPER format
//	f_, _ := time.Parse(layout, date)
//	d := f_.Format(layout)
//	return d
//}
//

//
//func PushBatchedPayloadIntoDB(output []structs.ExpandedPolygonStocksAggResponseParams, poolConfig *pgxpool.Config, batchSize int) {
//	var j int
//	numInserts := len(output)
//	iterations := (numInserts - 1) / batchSize
//	ctx := context.Background()
//	bar := progressbar.Default(int64(iterations), "Uploading...")
//
//	go func(batchSize int, numInserts int, bar *progressbar.ProgressBar) {
//		connPool, err := pgxpool.ConnectConfig(ctx, poolConfig)
//		if err != nil {
//			fmt.Println("Unable to create conn...", err)
//		}
//		defer connPool.Close()
//
//		batch := &pgx.Batch{}
//		wg := new(sync.WaitGroup)
//
//		go func(batchSize int, numInserts int, batch *pgx.Batch, wg *sync.WaitGroup, bar *progressbar.ProgressBar) {
//			defer wg.Done()
//			for i := 0; i < numInserts-1; i += batchSize {
//				wg.Add(1)
//				if ((numInserts - 1) - i) == 1 {
//					j = i + 1
//				} else {
//					j = i + batchSize
//				}
//
//				for k := range output[i:j] {
//					batch.Queue(db.PolygonStocksAggCandlesInsertTemplate,
//						output[k].Ticker,
//						output[k].Timespan,
//						output[k].Multiplier,
//						output[k].V,
//						output[k].Vw,
//						output[k].O,
//						output[k].C,
//						output[k].H,
//						output[k].L,
//						output[k].T)
//				}
//
//				// pull through the batch and exec each statement
//				br := connPool.SendBatch(context.Background(), batch)
//				for k := 0; k < (j - i); k++ {
//					_, err := br.Exec()
//					if err != nil {
//						fmt.Println("Unable to execute statement in batched queue: ", err)
//						os.Exit(1)
//					}
//				}
//
//				// Close this batch pool
//				var err = br.Close()
//				if err != nil {
//					fmt.Println("Unable to close batch: ", err)
//				}
//
//				err = bar.Add(1)
//				if err != nil {
//					fmt.Println("Something wrong with inserting batches bar ", err)
//				}
//			}
//		}(batchSize, numInserts, batch, wg, bar)
//
//		wg.Wait()
//	}(batchSize, numInserts, bar)
//}

//func main_old() {
//	var urls []*url.URL
//
//	// Read all the equities into a list, grab the length
//	equitiesList := config.ReadEquitiesList()
//
//	// Convert all dates to a PROPER format
//	from_ := ParseTime(from_)
//	to_ := ParseTime(to_)
//
//	// Make all urls, dont do it on the fly
//	// Set polygonIo cred
//	polygonApiKey := config.SetPolygonCred("me")
//	urls = MakeAllStocksAggsQueries(equitiesList, timespan, from_, to_, polygonApiKey)
//
//	// Set DB params and make a connection pool
//	postgresParams := new(config.DbParams)
//	err := config.SetDBParams(postgresParams, "postgres")
//	if err != nil {
//		fmt.Println(err)
//	}
//	postgresConnStr := fmt.Sprintf(
//		"postgres://%s:%s@%s:%s/%s",
//		postgresParams.User,
//		postgresParams.Password,
//		postgresParams.Host,
//		postgresParams.Port,
//		postgresParams.Dbname,
//	)

//poolConfig, err := pgxpool.ParseConfig(postgresConnStr)
//if err != nil {
//	fmt.Println("Unable to create config str...", err)
//}

//bar1 := progressbar.Default(int64(len(urls)), "Downloading...")
//c := MakeAllStocksAggsRequests(urls, bar1)
//
//bar2 := progressbar.Default(int64(len(urls)), "Flattening...")
////var bigPayload []structs.ExpandedPolygonStocksAggResponseParams
//
//for payload := range c {
//	output := db.PolygonStocksAggCandlesFlattenPayloadBeforeInsert(payload, timespan, multiplier, layout)
//
//	if len(output) > 0 {
//		PushGiantPayloadIntoDB(output, poolConfig)
//		//bigPayload = append(bigPayload, output...)
//	}
//
//	err = bar2.Add(1)
//	if err != nil {
//		fmt.Println("\nSomething wrong with bar2: ", err)
//	}
//}

//PushBatchedPayloadIntoDB(bigPayload, poolConfig, 5000)
//}
