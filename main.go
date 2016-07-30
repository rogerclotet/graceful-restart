package main

import "net/http"
import (
	"context"
	"fmt"
	"github.com/rogerclotet/graceful-restart/cqrs/argument"
	"github.com/rogerclotet/graceful-restart/cqrs/command"
	"github.com/rogerclotet/graceful-restart/cqrs/query"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type handled struct {
	mu *sync.RWMutex
	n  int
}

func (h *handled) increment() {
	h.mu.Lock()
	h.n++
	h.mu.Unlock()
}

func main() {
	h := handled{
		mu: &sync.RWMutex{},
	}

	commandRegistry, err := command.NewRegistry(
		command.NewRegisteredCommand("increment", incrementCommand(&h)),
	)
	if err != nil {
		log.Fatalf("could not create command registry: %v", err)
	}

	queryRegistry, err := query.NewRegistry(
		query.NewRegisteredQuery("handled_requests", handledRequestsQuery(&h)),
	)
	if err != nil {
		log.Fatalf("could not create query registry: %v", err)
	}

	cmdQueue := make(chan command.Command)
	cmdToHandle := make(chan command.Command)
	start := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	go commandQueue(ctx, cmdQueue, cmdToHandle, start)
	start <- struct{}{}

	var wg sync.WaitGroup
	go commandHandler(context.Background(), commandRegistry, cmdToHandle, &wg)
	defer wg.Wait()

	queries := make(chan query.Query)
	go queryHandler(context.Background(), queryRegistry, queries)

	http.Handle("/command", func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			name := r.URL.Query().Get("cmd")
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			cmdQueue <- command.New(name, argument.Arguments{}) // TODO parse arguments
		}
	}())

	http.Handle("/query", func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			name := r.URL.Query().Get("q")
			if name == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			q := query.New(name, argument.Arguments{}) // TODO parse arguments
			queries <- q

			response := <-q.Response()

			fmt.Fprint(w, response)
		}
	}())

	go func() {
		err = http.ListenAndServe(":8080", http.DefaultServeMux)
		if err != nil {
			log.Printf("error in ListenAndHandle: %v", err)
		}
	}()

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	<-stop

	cancel() // Cancel context passed to commandQueue
}

func commandQueue(ctx context.Context, commands chan command.Command, toHandle chan command.Command, start chan struct{}) {
	processing := false
	var queue []command.Command

	for {
		select {
		case <-start:
			for _, c := range queue {
				toHandle <- c
			}
			queue = []command.Command{}
			processing = true
		case c := <-commands:
			if processing {
				toHandle <- c
			} else {
				queue = append(queue, c)
			}
		case <-ctx.Done():
			processing = false
		}
	}
}

func commandHandler(ctx context.Context, r command.Registry, commands chan command.Command, wg *sync.WaitGroup) {
	for c := range commands {
		wg.Add(1)
		err := r.Handle(ctx, c)
		if err != nil {
			log.Printf("error handling command %s: %v", c.Name(), err)
		}
		wg.Done()
	}
}

func incrementCommand(h *handled) command.Handler {
	return func(_ context.Context, _ argument.Arguments) error {
		h.increment()

		return nil
	}
}

func queryHandler(ctx context.Context, r query.Registry, queries chan query.Query) {
	for q := range queries {
		res, err := r.Handle(ctx, q)
		if err != nil {
			log.Printf("error handling query %s: %v", q.Name(), err)
			q.Respond(nil)
			return
		}

		q.Respond(res)
	}
}

func handledRequestsQuery(h *handled) query.Handler {
	return func(_ context.Context, _ argument.Arguments) (interface{}, error) {
		h.mu.RLock()
		defer h.mu.RUnlock()

		return h.n, nil
	}
}
