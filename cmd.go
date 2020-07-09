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

	"github.com/grandcat/zeroconf"
	"github.com/jlaffaye/ftp"

	"goftp.io/server/core"
	"goftp.io/server/driver/file"
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

func configFtpServer(iName string) *core.Server {
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
	factory := &file.DriverFactory{
		RootPath: *servDir,
		Perm:     core.NewSimplePerm("user", "group"),
	}
	opts := &core.ServerOpts{
		Name:           iName,
		WelcomeMessage: "Welcom to the drop link FTP server",
		Factory:        factory,
		Port:           servPort,
		//Hostname: *host, //according to the doc, the default value is enough.
		Auth: &core.SimpleAuth{Name: *user, Password: *pass},
	}

	log.Printf("the serve dir is %s\n", *servDir)
	log.Printf("username %v, password %v", *user, *pass)
	return core.NewServer(opts)
}

func startFtpServer(ftpServer *core.Server) {

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
			//log.Println("receive finished message from the done channel")
			return
		default:
			//don't block it
		}
	}

	//log.Println("browse finished.")
	return
}

func uploadFile(host string, port int, file string) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// step 2: uploaded the specivied file
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second), ftp.DialWithDisabledEPSV(false))
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
		log.Fatal(err)
	}

	if err := c.Quit(); err != nil {
		log.Fatal(err)
	}

}

// inputDest show the list to the client, let them choose one by index
// return the client chosen index
func chooseInstance(ir []iRecord) (idx int) {
	fmt.Printf("%5s | %s\n", "index", "service name @ machine")

	for i, one := range ir {
		fmt.Printf("[%3d] | %s@%s\n", i, one.iName, one.hostname)
	}
	for {
		fmt.Printf("please use the index to choose.\n")
		_, err := fmt.Scanf("%d\n", &idx)
		if err != nil {
			fmt.Printf("wrong input! num=%d,err=%s\n", &idx, err)
			continue
		}
		if (idx <= len(ir)-1) && (idx >= 0) {
			break //right between 0:len([]iRecodd)-1
		} else {
			fmt.Printf("wrong index range, [0-%d] is valid\n", len(ir)-1)
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
		if len(ir) == 0 {
			fmt.Println("there is no droplocal service available.")
			return
		}
		idx := chooseInstance(ir)

		//step3: upload the file to the target machine

		//TODO goftp doesn't support EPSV for ipv6
		//just switch to ipv4 address for onetime upload
		target, err := getIPv4from(ir[idx].hostname)
		if err != nil {
			fmt.Printf("can't get ipv4 address: %s, fall back to use hostname: %s\n", err, ir[idx].hostname)
			target = ir[idx].hostname
		}
		uploadFile(target, ir[idx].port, *files)
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
