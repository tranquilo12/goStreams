package db

// https://api.polygonio/v2/aggs/{stocksTicker}/range/{multiplier}/{timespan}/{from}/{to}

// UploadAggToS3 s3 path /aggs/{insert_date}/{timespan_multiplier}/{trade_date}/{stocksTickers}/data.tape
//func UploadAggToS3(svc *s3.S3, url *url.URL, response *http.Response) error {
//
//	bucket := "polygonio-all"
//
//	// get today's date
//	today := time.Now().Format("2006-01-02")
//
//	// get all the necessary details from the URL
//	details := getDetailsFromUrl(url)
//
//	// Generate the S3 key that will be inserted into the bucket
//	s3Key := fmt.Sprintf("%s/%s/%s-%s/%s/%s", "aggs", today, details.Timespan, details.Multiplier, details.From, "data.tape")
//
//	// Make Key
//	res, err := keyExists(svc, "polygonio-all", s3Key)
//	if err != nil {
//		panic(err)
//	}
//
//	// Marshal target to bytes
//	target := new(structs.AggregatesBarsResponse)
//	err = json.NewDecoder(response.Body).Decode(&target)
//	taskBytes, err := json.Marshal(target)
//	if err != nil {
//		fmt.Println("Error retrieving URL: ", err)
//	}
//
//	if !res {
//		input := &s3.PutObjectInput{
//			Body:   aws.ReadSeekCloser(strings.NewReader(taskBytes)),
//			Bucket: aws.String(bucket),
//			Key:    aws.String(s3Key),
//			ContentLength: aws.Int64(response.ContentLength),
//			ContentType:
//		}
//		resp, err := svc.PutObject(input)
//		if err != nil {
//			panic(err)
//		}
//
//		fmt.Printf("inserted response: %s\n", awsutil.StringValue(resp))
//
//	}
//
//	return nil
//}
