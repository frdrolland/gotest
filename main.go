//
// gotest project main.go
//
package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/golang/glog"
)

const DEFAULT_PID_FILE = "daemon.pid"

const DEFAULT_LISTENING_INTERFACE = ":8080"

// Command-line flags.
var (
	helpFlag         = flag.Bool("h", false, "Show this help")
	pidFlag          = flag.String("p", DEFAULT_PID_FILE, "File to store PID when running")
	httpAddr         = flag.String("i", DEFAULT_LISTENING_INTERFACE, "Listen interface")
	timeout          = 5 * time.Second
	ErrNotRunning    = errors.New("process not running")
	processId        = os.Getpid()
	doCleanupPidFile = false
	wg               sync.WaitGroup
)

/**
 * All requests.
 */
func allHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[debug] Adding header")
	w.Header().Add("Server", "gotest/0.1.1")
}

/**
 * Healthcheck implementation.
 */
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

/**
 * Healthcheck implementation.
 */
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ready"))
}

/**
 * Healthcheck implementation.
 */
func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("server is shutting down..."))
	wg.Done()
}

/**
 * Implementation of CRUD resource
 */
func crudHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not yet implemented", http.StatusInternalServerError)
	return
}

// getProcess gets a Process from a pid and checks that the
// process is actually running. If the process
// is not running, then getProcess returns a nil
// Process and the error ErrNotRunning.
func getProcess(pid int) (*os.Process, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	// try to check if the process is actually running by sending
	// it signal 0.
	err = p.Signal(syscall.Signal(0))
	if err == nil {
		return p, nil
	}
	if err == syscall.ESRCH {
		return nil, ErrNotRunning
	}
	return nil, errors.New("server running but inaccessible")
}

/**
 * Checks an error and panic if not nil.
 */
func check(e error) {
	if e != nil {
		panic(e)
	}
}

/**
 *
 */
func cleanup() {
	if doCleanupPidFile {
		log.Println("Cleaning up pid file")
		err := os.Remove(*pidFlag)
		check(err)
		log.Println("pid file has been removed")
	} else {
		log.Println("[WARN] won't clean up PID file")
	}
}

/**
 * Main function
 */
func main() {
	flag.Parse()
	if true == *helpFlag {
		flag.Usage()
		os.Exit(-1)
	}

	glog.Info("Starting - pid = ", processId)

	bPid := []byte(strconv.Itoa(processId))

	f, err := os.OpenFile(*pidFlag, os.O_CREATE|os.O_RDWR|os.O_SYNC, 0666)
	check(err)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := strings.Trim(scanner.Text(), "")
		log.Printf("str = %s", str)
		ppid, err := strconv.Atoi(str)
		check(err)

		if ppid > 0 {
			killErr := syscall.Kill(ppid, syscall.Signal(0))
			if nil == killErr {
				log.Printf("A process is already running with pid %d : exiting\n", ppid)
				os.Exit(2)
			}
		}
		log.Printf("PID file not empty but old process (pid = %d) is down... maybe a crash occured ?\n", ppid)
		err = f.Truncate(0)
		check(err)
	}

	log.Println("PID = ", string(bPid))
	f.WriteString(string(bPid))
	f.WriteString("\n")
	f.Sync()
	doCleanupPidFile = true

	// Catch signals to cleanup before exiting
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM|syscall.SIGKILL)
	go func() {
		<-c
		cleanup()
		log.Println("[exit]")
		os.Exit(0)
	}()

	// ========================================
	// HTTP Server start
	wg.Add(1)

	// Prepare handler for all URIs
	http.HandleFunc("*", allHandler)
	http.HandleFunc("/healthcheck", healthCheckHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/crud/", crudHandler)
	http.HandleFunc("/admin/shutdown", shutdownHandler)
	//	http.HandleFunc("/admin/set/standby", setStandbyHandler)
	//	http.HandleFunc("/admin/set/running", setRunningHandler)

	// Start HTTP server in another thread
	log.Println("Starting HTTP server on adress : ", *httpAddr)
	http.Handle("/static/", http.FileServer(http.Dir("./static")))
	go log.Fatal(http.ListenAndServe(*httpAddr, nil))
	log.Println("HTTP server started")

	// Wait for termination
	wg.Wait()
}
