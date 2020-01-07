package main

import (
    "context"
    "flag"
    "fmt"
    api "github.com/synerex/synerex_api"
    pbase "github.com/synerex/synerex_proto"
    pcounter "github.com/synerex/proto_pcounter"    
    "github.com/synerex/synerex_sxutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

    "log"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// datastore provider provides Datastore Service.

type DataStore interface{
    store(str string)
}

var (
	nodesrv    = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	local		= flag.String("local","","Local Synerex Server")
    mu         sync.Mutex
    version = "0.01"
    baseDir = "store"
    dataDir string
    ds DataStore
)

func init(){
    var err error
    dataDir, err =os.Getwd()
    if err != nil {
        fmt.Printf("Can't obtain current wd")
    }
    dataDir =filepath.ToSlash(dataDir) + "/" + baseDir
    ds = &FileSystemDataStore{
        storeDir:dataDir,
    }
}

type FileSystemDataStore struct{
    storeDir string
    storeFile *os.File
    todayStr string
}

// open file with today info
func (fs *FileSystemDataStore)store(str string){
    const layout = "2006-01-02"
    day := time.Now()
    todayStr := day.Format(layout)+".csv"
    if fs.todayStr != "" && fs.todayStr != todayStr {
        fs.storeFile.Close()
        fs.storeFile = nil
    }
    if fs.storeFile == nil {
        _, er := os.Stat(fs.storeDir)
        if er != nil {// create dir
            er = os.MkdirAll(fs.storeDir, 0777)
            if er != nil {
                fmt.Printf("Can't make dir '%s'.",fs.storeDir)
                return
            }
        }
        fs.todayStr = todayStr
        file, err := os.OpenFile(filepath.FromSlash(fs.storeDir+"/"+todayStr),os.O_RDWR| os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
            fmt.Printf("Can't open file '%s'",todayStr)
            return
        }
        fs.storeFile =file
    }
    fs.storeFile.WriteString(str+"\n")
}


func supplyPCounterCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {

    pc := &pcounter.PCounter{}

    err := proto.Unmarshal(sp.Cdata.Entity, pc)
	if err == nil{ // get Pcounter
		ts0 := ptypes.TimestampString(pc.Ts)
		ld := fmt.Sprintf("%s,%s,%s,%s,%s",ts0,pc.Hostname,pc.Mac,pc.Ip,pc.IpVpn)
		ds.store(ld)
		for _, ev := range pc.Data {
			ts := ptypes.TimestampString(ev.Ts)
			line := fmt.Sprintf("%s,%s,%d,%s,%s,",ts,pc.DeviceId,ev.Seq,ev.Typ,ev.Id)
			switch (ev.Typ){
			case "counter":
				line = line + fmt.Sprintf("%s,%d",ev.Dir,ev.Height)
			case "fillLevel":
				line = line + fmt.Sprintf("%d",ev.FillLevel)
			case "dwellTime":
				tsex := ptypes.TimestampString(ev.TsExit)
				line = line + fmt.Sprintf("%f,%f,%s,%d,%d",ev.DwellTime,ev.ExpDwellTime,tsex,ev.ObjectId,ev.Height)			
			}
			ds.store(line)
		}
    }
}

func subscribePCounterSupply(client *sxutil.SXServiceClient) {
    ctx := context.Background() //
    client.SubscribeSupply(ctx, supplyPCounterCallback)
    log.Fatal("Error on subscribe")
}


func main() {
    flag.Parse()
    go sxutil.HandleSigInt()    
    sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)

    channelTypes := []uint32{pbase.PEOPLE_COUNTER_SVC}

    srv, rerr := sxutil.RegisterNode(*nodesrv, "PCouterStore", channelTypes, nil)

    if rerr != nil {
        log.Fatal("Can't register node:",rerr)
	}
	if *local != ""{// quick hack for AWS local network
		srv = *local
	}
    log.Printf("Connecting SynerexServer at [%s]", srv)

    wg := sync.WaitGroup{} // for syncing other goroutines

    client := sxutil.GrpcConnectServer(srv)

    if client == nil {
        log.Fatal("Can't connect Synerex Server")
    }else{
        log.Print("Connecting SynerexServer")
    }
    
    pc_client := sxutil.NewSXServiceClient(client, pbase.PEOPLE_COUNTER_SVC,"{Client:PcountStore}")

    wg.Add(1)
    log.Print("Subscribe Supply")
    go subscribePCounterSupply(pc_client)

    wg.Wait()

}


