package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/rogerclotet/graceful-restart/cqrs/argument"
	"github.com/rogerclotet/graceful-restart/cqrs/command"
	"github.com/rogerclotet/graceful-restart/cqrs/query"
)

const inheritedFileDescriptor = 3

// Data is the app data to be stored in a snapshot
type Data struct {
	mu *sync.RWMutex
	N  int
}

var wg sync.WaitGroup

func (d *Data) increment() {
	d.mu.Lock()
	d.N++
	d.mu.Unlock()
}

func main() {
	fmt.Printf("hi! I'm %d\n", os.Getpid())

	var graceful bool
	flag.BoolVar(&graceful, "graceful", false, "restarting gracefully, internal use only")
	flag.Parse()

	commands := make(chan interface{})
	cmdToHandle := make(chan interface{})
	queries := make(chan interface{})
	qToHandle := make(chan interface{})
	processCommands := make(chan bool)
	processQueries := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())
	go queue(ctx, commands, cmdToHandle, processCommands)
	go queue(ctx, queries, qToHandle, processQueries)

	http.Handle("/command", func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uq := r.URL.Query()
			args := argsFromURLQuery(uq)
			name, err := args.GetString("cmd")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			commands <- command.New(name, args)
		}
	}())

	http.Handle("/query", func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uq := r.URL.Query()
			args := argsFromURLQuery(uq)
			name, err := args.GetString("q")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			q := query.New(name, args)
			queries <- q

			qr := <-q.Response()
			if qr.Err() != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			fmt.Fprint(w, qr.Response())
		}
	}())

	var l net.Listener
	var err error
	if graceful {
		f := os.NewFile(inheritedFileDescriptor, "")
		l, err = net.FileListener(f)
		if err != nil {
			log.Fatalf("could not listen to inherited socket: %v", err)
		}
	} else {
		l, err = net.Listen("tcp", ":8080")
		if err != nil {
			log.Fatalf("could not listen to TCP port 8080: %v", err)
		}
	}

	server := &http.Server{
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	netListener := newGracefulListener(l)

	go func() {
		_ = server.Serve(netListener)
	}()

	d := restoreSnapshot()

	commandRegistry, err := command.NewRegistry(
		command.NewRegisteredCommand("increment", incrementCommand(&d)),
	)
	if err != nil {
		log.Fatalf("could not create command registry: %v", err)
	}

	queryRegistry, err := query.NewRegistry(
		query.NewRegisteredQuery("handled_commands", handledCommandsQuery(&d)),
	)
	if err != nil {
		log.Fatalf("could not create query registry: %v", err)
	}

	var wg sync.WaitGroup
	handlerCtx := context.Background()
	go commandHandler(handlerCtx, commandRegistry, cmdToHandle, &wg)
	go queryHandler(handlerCtx, queryRegistry, qToHandle, &wg)
	defer wg.Wait()

	processQueries <- true
	processCommands <- true

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	restart := make(chan os.Signal)
	signal.Notify(restart, syscall.SIGHUP)

	select {
	case <-stop:
		cancel()
		wg.Wait()
		d.takeSnapshot()
	case <-restart:
		cancel()
		wg.Wait()
		d.takeSnapshot()
		startFork(netListener)
	}
}

func (d Data) takeSnapshot() {
	path, _ := filepath.Abs("./data.gob")
	file, err := os.Create(path)
	if err != nil {
		log.Printf("failed to create snapshot: %v", err)
	}

	_ = gob.NewEncoder(file).Encode(d)
}

func restoreSnapshot() Data {
	path, _ := filepath.Abs("./data.gob")
	file, err := os.Open(path)
	d := Data{
		mu: &sync.RWMutex{},
	}
	if err == nil {
		_ = gob.NewDecoder(file).Decode(&d)
	}
	return d
}

func startFork(l *gracefulListener) {
	file := l.File()

	args := []string{}
	if len(os.Args) > 1 {
		args = append(args, os.Args[1:]...)
	}
	args = append(args, "-graceful") // TODO avoid repeating flag after second reload
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{file}

	err := cmd.Start()
	if err != nil {
		log.Fatalf("gracefulRestart: Failed to launch, error: %v", err)
	}
}

func argsFromURLQuery(query url.Values) argument.Arguments {
	var args argument.Arguments
	for k, v := range query {
		args[k] = argument.New(v)
	}
	return args
}

func queue(ctx context.Context, in chan interface{}, out chan interface{}, process chan bool) {
	processing := false
	var queue []interface{}

	for {
		select {
		case p := <-process:
			if p {
				for _, elem := range queue {
					out <- elem
				}
				queue = []interface{}{}
			}
			processing = p
		case elem := <-in:
			if processing {
				out <- elem
			} else {
				queue = append(queue, elem)
			}
		case <-ctx.Done():
			return
		}
	}
}

func commandHandler(ctx context.Context, r command.Registry, commands chan interface{}, wg *sync.WaitGroup) {
	for receivedCommand := range commands {
		c, ok := receivedCommand.(command.Command)
		if !ok {
			log.Printf("received %v in command handler", receivedCommand)
			continue
		}

		wg.Add(1)
		err := r.Handle(ctx, c)
		if err != nil {
			log.Printf("error handling command %s: %v", c.Name(), err)
		}
		wg.Done()
	}
}

func incrementCommand(d *Data) command.Handler {
	return func(_ context.Context, _ argument.Arguments) error {
		d.increment()

		return nil
	}
}

func queryHandler(ctx context.Context, r query.Registry, queries chan interface{}, wg *sync.WaitGroup) {
	for receivedQuery := range queries {
		q, ok := receivedQuery.(query.Query)
		if !ok {
			log.Printf("received %v in query handler", receivedQuery)
			continue
		}

		wg.Add(1)
		res, err := r.Handle(ctx, q)
		qr := query.NewResponse(res, err)
		if err != nil {
			log.Printf("error handling query %s: %v", q.Name(), err)
		}
		q.Respond(qr)
		wg.Done()
	}
}

func handledCommandsQuery(d *Data) query.Handler {
	return func(_ context.Context, _ argument.Arguments) (interface{}, error) {
		d.mu.RLock()
		defer d.mu.RUnlock()

		return d.N, nil
	}
}

type gracefulListener struct {
	net.Listener
	stop chan error
}

func newGracefulListener(l net.Listener) (gl *gracefulListener) {
	gl = &gracefulListener{Listener: l, stop: make(chan error)}
	go func() {
		_ = <-gl.stop
		gl.stop <- gl.Listener.Close()
	}()
	return
}

func (gl *gracefulListener) Accept() (net.Conn, error) {
	c, err := gl.Listener.Accept()
	if err != nil {
		return c, err
	}

	c = gracefulConnection{Conn: c}

	wg.Add(1)
	return c, err
}

func (gl *gracefulListener) Close() error {
	gl.stop <- nil
	return <-gl.stop
}

func (gl *gracefulListener) File() *os.File {
	tcpListener := gl.Listener.(*net.TCPListener)
	f, err := tcpListener.File()
	if err != nil {
		log.Fatalf("could not get file descriptor for TCP socket: %v", err)
	}
	return f
}

type gracefulConnection struct {
	net.Conn
}

func (w gracefulConnection) Close() error {
	wg.Done()
	return w.Conn.Close()
}
