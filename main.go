package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to terminate server: %v", err)
	}
}

func run(ctx context.Context) error {
	s := &http.Server{
		Addr: ":18080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("failed to close: %+v", err)
			return err
		}
		return nil
	})

	<-ctx.Done()
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}

	return eg.Wait()
}

// 채널에서 알림이 온 경우는 다음과 같은 순서로 run 함수가 종료된다.

// - ctx.Done() 다음 줄의 *http.Server.Shutdown 메서드가 실행된다.
// - 다른 고루틴에서 실행되던 *http.Server.ListenAndServe 메서드가 종료된다.
// - 다른 고루틴에서 실행되던 익명 함수(func() error)가 종료된다.
// - run 함수의 마지막 부분에서 다른 고루틴이 종료되는 것을 기다리던 *errgroup.Group.Wait 메서드가 종료된다.
// - 다른 고루틴에서 실행되던 익명 함수(func() error)의 반환값이 run 함수의 반환값이 된다.
