package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"therealbroker/api/api/proto"
	"time"
)

func main() {
	var wg sync.WaitGroup

	dial, err := grpc.Dial(":8001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("could not connect!\n", err)
	}

	defer func(dial *grpc.ClientConn) {
		err := dial.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(dial)

	ticker := time.NewTicker(1 * time.Microsecond)

	go func() {
		for {
			//i := 0
			select {
			case <-ticker.C:
				wg.Add(1)
				go func() {
					defer wg.Done()

					client := proto.NewBrokerClient(dial)
					ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
					defer cancel()

					pubReq := proto.PublishRequest{
						Subject:           "hi",
						Body:              []byte("this is a test"),
						ExpirationSeconds: 20,
					}
					pubRes, err := client.Publish(ctx, &pubReq)
					if err != nil {
						log.Fatal(err)
					} else {
						fmt.Println(pubRes.Id)
					}
				}()
			default:

			}
			//i++
		}
	}()

	time.Sleep(1800 * time.Second)
	ticker.Stop()
	//for i := 0; i < 1000; i++ {
	//	wg.Add(1)
	//
	//	go func() {
	//		defer wg.Done()
	//
	//		subReq := proto.SubscribeRequest{
	//			Subject: "hi" + string(i%23),
	//		}
	//
	//		_, err = client.Subscribe(context.Background(), &subReq)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//	}()
	//}

	//for i := 0; i < 1000; i++ {
	//	wg.Add(1)
	//
	//	go func() {
	//		defer wg.Done()
	//
	//		fetchReq := proto.FetchRequest{
	//			Id: int32(2),
	//			Subject: "hi" + string(i%23),
	//		}
	//
	//	}()
	//}

	wg.Wait()
}
