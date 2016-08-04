// wdtest project main.go
package main

import (
	//"fmt"
	"log"
	"net/http"
	//	"os"
	//	"syscall"
	"sync"
	"time"
)

var i, j int = 1, 2
var flagvar int

/**
 *
 */
func helloHandler(w http.ResponseWriter, r *http.Request) {
	r.Header.Add("test", "fro")
	w.Write([]byte("OK"))
}

/**
 * Main method
 */
func main() {

	log.Println("The time is", time.Now())
	var wg sync.WaitGroup
	wg.Add(1)

	http.HandleFunc("/healthcheck", helloHandler)

	log.Println("Starting HTTP server on port 8080")
	go http.ListenAndServe(":8080", nil)
	log.Println("HTTP server started")
	/*
		sys_attr.Ptrace = true
		proc_attr.Sys = &sys_attr

		proc_attr.Files = []uintptr{uintptr(syscall.Stdin),
			uintptr(syscall.Stdout), uintptr(syscall.Stderr)}

		pid, err := syscall.ForkExec("wdtest", []string{"wdtest"}, &proc_attr)

		var proc_attr syscall.ProcAttr
		var sys_attr syscall.SysProcAttr

		if err != nil {
			//		log.Fatal("Error '", err, "' occured on fork of process")
			return err
		}

		fmt.Println("Process PID[22689] forked successfully to PID[22690]")
	*/
	wg.Wait()
	log.Println("[exit]")
}
