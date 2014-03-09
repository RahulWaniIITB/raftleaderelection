package main
import (

	"fmt"
	"time"
	"os"
	"strconv"
	"sync"
	"net/rpc"
	"net"
	"Raft"


)


var r *Raft.Raft_Implementer




func main() {

	//exit_main=make(chan *bool)

	id,_:=strconv.Atoi(os.Args[1])

	number_of_servers:=5
	size_of_in_chan:=20
	size_of_out_chan:=20

	parent_dir:=""+os.Args[3]
	fnm:=parent_dir+"/Raft/configuration"+strconv.Itoa(id)+".xml"

	//fmt.Println(""+parent_dir+":"+fnm)

	r=Raft.NewRaft(id, size_of_in_chan,size_of_out_chan,100*time.Millisecond,fnm,parent_dir)

	wg:=new(sync.WaitGroup)
	wg.Add(2)
	go r.S.Send(number_of_servers,wg)
	go r.S.Receive(number_of_servers,wg)
	go r.Recv_From_Inbox()
	go r.RaftController()
	//go raft.Close()
	fmt.Println("\nIn main")


	raft_rpc := new(Raft.RaftRPC)
	rpc.Register(raft_rpc)
	///rpc.HandleHTTP()

	l, e := net.Listen("tcp", ":"+os.Args[2])


	if e != nil {
		//log.Fatal("listen error:", e)
		fmt.Printf("\nError in listening")
	}
	//go http.Serve(l, nil)
	//fmt.Println("Waiting")
	go func() {
		for {
			conn, errconn := l.Accept()
			if errconn!=nil {
				fmt.Printf("\nConnection error")
			}
			go rpc.ServeConn(conn)
		}
	}()


	wg.Wait()
	fmt.Println("Exiting")



}
func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
