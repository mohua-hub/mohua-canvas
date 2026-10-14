package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	httpapi "github.com/mohua-hub/mohua-canvas/internal/adapters/inbound/http"
	"github.com/mohua-hub/mohua-canvas/internal/bootstrap"
)
func main() {
	log.SetFlags(0)
	dataDir,token:=os.Getenv("MOHUA_DATA_DIR"),os.Getenv("MOHUA_SERVICE_TOKEN")
	key,err:=base64.StdEncoding.DecodeString(os.Getenv("MOHUA_VAULT_KEY"));_ = os.Unsetenv("MOHUA_VAULT_KEY")
	tokenBytes,tokenErr:=hex.DecodeString(token);_ = os.Unsetenv("MOHUA_SERVICE_TOKEN")
	if err!=nil || len(key)!=32 || tokenErr!=nil || len(tokenBytes)!=32 || !filepath.IsAbs(dataDir) { log.Fatal("SERVICE_CONFIGURATION_INVALID") }
	configPath:=os.Getenv("MOHUA_PLUGIN_CONFIG");if configPath==""{configPath="config/plugins.json"}
	runtime,err:=bootstrap.New(dataDir,configPath,key);for i:=range key{key[i]=0}
	if err!=nil{log.Fatal("SERVICE_INITIALIZATION_FAILED")};defer runtime.Close()
	listener,err:=net.Listen("tcp","127.0.0.1:0");if err!=nil{log.Fatal("SERVICE_LISTEN_FAILED")}
	handler:=httpapi.New(httpapi.Services{Projects:runtime.Projects,Channels:runtime.Channels,Assets:runtime.Assets,Jobs:runtime.Jobs,Library:runtime.Library,Plugins:runtime.Plugins},token)
	server:=&http.Server{Handler:handler,ReadHeaderTimeout:10*time.Second,ReadTimeout:30*time.Second,WriteTimeout:2*time.Minute,IdleTimeout:30*time.Second,MaxHeaderBytes:16<<10}
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer stop()
	go func(){_,_=io.Copy(io.Discard,os.Stdin);stop()}()
	done:=make(chan error,1);go func(){done<-server.Serve(listener)}()
	fmt.Printf("MOHUA_READY=%d\n",listener.Addr().(*net.TCPAddr).Port)
	select{case <-ctx.Done():case err:=<-done:if err!=http.ErrServerClosed{log.Print("SERVICE_STOPPED")}}
	shutdown,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
	_=server.Shutdown(shutdown)
}
