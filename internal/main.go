package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MapleGu/dns/dns"
	"github.com/MapleGu/dns/store"
	_ "github.com/jpfuentes2/go-env/autoload"
)

var (
	rwDirPath = os.Getenv("RWDirPath")
	addr      = os.Getenv("DNSADDR")
)

func main() {
	book := store.NewStore(rwDirPath)
	book.Load()

	dnsService := dns.NewService(book)
	go dnsService.Start(addr)

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGABRT, syscall.SIGHUP)
	<-sig
}
