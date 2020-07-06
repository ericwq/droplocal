package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	ouser "os/user"
	"path/filepath"
	"syscall"
	"time"

	fdriver "github.com/goftp/file-driver"
	gserver "github.com/goftp/server"
	"github.com/grandcat/zeroconf"
	"github.com/jlaffaye/ftp"
)

const servNamePrefix = "Drop Local"
const servType = "_droplocal._tcp"
const servPort = 2121
const servDomain = "local."

var (

	//droplocal client configuration
	files = flag.String("f", "", "source file for upload")

	//used by zeroconf client
	servMode = flag.Bool("s", false, "run in server mode")

	//server configuration servDir default set to User's home dir see func newFtpServer()
	servDir = flag.String("d", "", "dir to serve file uploaded")
	user    = flag.String("u", "admin", "Username for login")
	pass    = flag.String("p", "password", "Password for login")
)

// advertise register the service with pre-defined servName, ServType, ServDomain, ServPort
// the iName is the instace name, it should be unique in the local network
func advertiseWith(iName string) *zeroconf.Server {
	// prepare to advertise the droplocal service
	// extra information about our service
	meta := []string{
		"version=0.1.0",
		"hello=droplocal", //TODO reserved for next phase
	}

	log.Printf("using the name:%s\n", iName)
	service, err := zeroconf.Register(
		iName,      // service instance name
		servType,   // service type and protocl
		servDomain, // service domain default is `loacl.`
		servPort,   // service port
		meta,       // service metadata
		//nil,        // register on all network interfaces
		filterInterface(),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("advertise the %s on local link", iName)

	return service
}

func configFtpServer(iName string) *gserver.Server {
	// the server needs a dir to save the download files
	if *servDir == "" {
		//using the user's home directory as the default serve directory
		currentUser, err := ouser.Current()
		if err != nil {
			log.Printf("try to get the current user's home directory %s\n", err)
		}
		*servDir = currentUser.HomeDir
	}

	// prepare to start the droplocal services
	factory := &fdriver.FileDriverFactory{
		RootPath: *servDir,
		Perm:     gserver.NewSimplePerm("user", "group"),
	}
	opts := &gserver.ServerOpts{
		Name:    iName,
		Factory: factory,
		Port:    servPort,
		//Hostname: *host, //according to the doc, the default value is enough.
		Auth: &gserver.SimpleAuth{Name: *user, Password: *pass},
	}

	log.Printf("the serve dir is %s\n", *servDir)
	log.Printf("username %v, password %v", *user, *pass)
	return gserver.NewServer(opts)
}

func startFtpServer(ftpServer *gserver.Server) {

	// start the droplocal service
	err := ftpServer.ListenAndServe()
	//main goroutine will waiting here for the incomming request
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}

type iRecord struct {
	iName    string
	hostname string
	port     int
}

// listInstance list the available instance with specified servName and servDomain
func listInstance(servType, servDomain string, timeoutSecond int) (records []iRecord, err error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Printf("Failed to initialize resolver:%s", err.Error())
		return
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSecond))
	defer cancel()
	err = resolver.Browse(ctx, servType, servDomain, entries)
	if err != nil {
		log.Printf("Failed to browse:%s", err.Error())
		return
	}

	//log.Println("browse the local link to find the available instance")
	//TODO the timeout setting above may cause some problem for a crowded local link
	//it will stop the range loop earlier than expected
	for e := range entries {
		//log.Printf("receive the entry %s %s:%d\n", e.Instance, e.HostName, e.Port)

		one := iRecord{iName: e.Instance, hostname: e.HostName, port: e.Port}
		records = append(records, one)
		//log.Printf("add record: %v\n", one)

		select {
		case <-ctx.Done():
			//ready to finish?
			log.Println("receive finished message from the done channel")
			return
		default:
			//don't block it
		}
	}

	//log.Println("browse finished.")
	return
}

/*
func browseIPandPort(instance string) (string, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(3))
	defer cancel()
	err = resolver.Browse(ctx, servType, servDomain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	log.Println("prepare to receive the browsed entry")
	//TODO the timeout setting above may cause some problem for a crowded local link
	//it will stop the range loop earlier than expected
	for e := range entries {
		log.Printf("receive the entry %s:%d\n", e.HostName, e.Port)

		// TODO simple solution: replace the `\ ` escape with ` `
		// consider standard DNS escape convention
		iName := strings.ReplaceAll(e.Instance, "\\ ", " ")
		log.Printf("entry: %s, instance:%s\n", iName, instance)

		if strings.HasPrefix(iName, instance) {
			//TODO returned when find the first match, consider the more instance case
			log.Printf("got matched instance:%s\n", iName)
			return fmt.Sprintf("%s:%d", e.HostName, e.Port), nil
		}

		select {
		case <-ctx.Done():
			//ready to finish?
			log.Println("receive finished message from the done channel")
			return "", errors.New("zeroconf browse context is done")
		default:
			//don't block it
		}
	}

	//deal with the case we didn't get the result
	return "", errors.New("Not find instance")
}

func uploadFile(instance, file string) {
	// step 1: lookup service name to get the hostname:ip and port
	log.Printf("the instance: %s, the file: %s\n", instance, file)
	addr, err := browseIPandPort(instance)
	if err != nil {
		log.Printf("find the instance error: %s\n", err)
		return
	}

	uploadFile(addr, file)
}
*/
/*
func uploadFile3(host string, port int, file string) {

	addr := fmt.Sprintf("%s:%d", host, port)
	uploadFile(addr, file)
}
*/
func uploadFile(host string, port int, file string) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// step 2: uploaded the specivied file
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login(*user, *pass)
	if err != nil {
		log.Fatal(err)
	}

	// open the specified file
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer f.Close() // we ensure close to avoid leaks

	_, fi := filepath.Split(file)
	// upload it
	err = c.Stor(fi, f)
	if err != nil {
		panic(err)
	}

	if err := c.Quit(); err != nil {
		log.Fatal(err)
	}

}

// inputDest show the list to the client, let them choose one by index
// return the client chosen index
func chooseInstance(ir []iRecord) (idx int) {
	for {
		fmt.Printf("%5s |  %s\n", "index", "service name @ machine")

		for i, one := range ir {
			fmt.Printf("[%3d] | %s@%s\n", i, one.iName, one.hostname)
		}
		fmt.Println("please choose the destination, please use the index to choose.")
		fmt.Scanf("%d\n", &idx)
		if (idx <= len(ir)-1) && (idx >= 0) {
			break //right between 0:len([]iRecodd)-1
		}
	}

	return
}

//if it's a regular file. return it. otherwise return ""
func checkFile(file string) bool {
	fileInfo, err := os.Lstat(file)
	if err != nil {
		log.Fatal(err)
	}

	supported := false
	//fmt.Printf("permissions: %#o\n", fileInfo.Mode().Perm()) // 0400, 0777, etc.
	switch mode := fileInfo.Mode(); {
	case mode.IsRegular():
		supported = true
	case mode.IsDir():
		fmt.Println("don't support directory!")
	case mode&os.ModeSymlink != 0:
		fmt.Println("don't support symbolic link!")
	case mode&os.ModeNamedPipe != 0:
		fmt.Println("don't support named pipe!")
	}

	return supported
}

func main() {
	flag.Parse()

	if *servMode == false {

		//step1: check the files first
		if len(*files) == 0 {
			fmt.Println("you should specified a file with -f")
			fmt.Printf("try %s -h\n", os.Args[0])
			return
		}
		if !checkFile(*files) {
			return
		}

		//step2: let user pick up the instance from the list
		ir, err := listInstance(servType, servDomain, 3)
		if err != nil {
			fmt.Printf("when pick up instance, there is an error: %s\n", err)
		}
		idx := chooseInstance(ir)

		//step3: upload the file to the target machine
		uploadFile(ir[idx].hostname, ir[idx].port, *files)
		fmt.Println("mission accomplished!")

	} else {
		// important lesson to learn:
		// Deferred Functions Are Not Always Guaranteed To Executed
		// such as the following case: received os.Interrupt or syscall.SIGTERM

		iName := instanceNameFactory() //TODO consider save the iName for future use
		ftpServer := configFtpServer(iName)
		go startFtpServer(ftpServer)
		// defer ftpServer.Shutdown()
		servAd := advertiseWith(iName)
		// defer service.Shutdown()

		//waiting for the os.Interrupt or syscall.SIGTERM
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sig:
			// shutdown the service when received the signal
			servAd.Shutdown()
			ftpServer.Shutdown()
			log.Println("shutting down...")
		}
	}
}

/*
func findInstance() {
	waitTime := 2
	// Discover all services on the network (e.g. _droplocal._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		//print the service on local link
		for entry := range results {
			fmt.Printf("found \"%[1]s\"\n", entry.Instance)
			log.Printf("ServiceName = %s\n", entry.ServiceName())
			log.Printf("HostName = %s\n", entry.HostName)
			log.Printf("Port = %v\n\n", entry.Port)
		}
		fmt.Println("\nNo more droplocal service.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(waitTime))
	defer cancel()
	err = resolver.Browse(ctx, servType, servDomain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)
}
*/
