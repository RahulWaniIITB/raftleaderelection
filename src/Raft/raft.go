package Raft
import (
	"encoding/xml"
	"fmt"
	"time"
	"os"
	"strconv"
	"bytes"
	"encoding/gob"
	//"log"
	//"sync"
	//"net/rpc"
	//"net"
	"cluster"
)







var raft Raft_Implementer

type RaftRPC int

type Int int
func (rpc *RaftRPC) Phase(args *Int,reply *int) error {
	*reply=raft.phase
	return nil
}
func (rpc *RaftRPC) Exit(args *Int,reply *int) error {
	//os.Exit(1)
	fmt.Printf("\nIn exit")
	//var t *time.Timer
	//var f func()

	/*f = func() {
		b:=true
		exit_main<-&b
	}*/

	time.AfterFunc( 5* time.Second, func(){os.Exit(2)})

	return nil

}






type Raft interface {
	Term()     int
	isLeader() bool
}

type Raft_Implementer struct {

	S cluster.ServerMain


	pid int
	ip string
	port int
	dir string
	timeout int
	serv []Raftserver
	own_index int

	curr_term 			int
	//bool_isleader 			bool
	heart_beat_frequency 		time.Duration

	chan_of_follower chan *Raft_Msg_Envelope
	chan_of_leader chan *Raft_Msg_Envelope
	chan_of_candidate chan *Raft_Msg_Envelope
	chan_exit_follower chan *bool
	chan_exit_leader chan *bool
	chan_exit_candidate chan *bool
	chan_exit_inbox chan *bool

	chan_exit_raft_controller chan *bool


	phase 				int
}

func (r *Raft_Implementer) Term() int {

	return r.curr_term
}


func (r *Raft_Implementer) isLeader() bool {

	if r.phase==1 {
		return true
	} else {
		return false
	}
	return false
}





type Raft_Msg_Envelope struct {
	Type_of_raft_msg int
	Pid int
	Term int
	Actual_msg interface {}
}



func (r *Raft_Implementer) createMsgToSend(pid int,msgId int,msgs []byte) cluster.Envelope {
	e:=cluster.Envelope{ Pid:pid, MsgId:msgId, Msg:msgs }
	return e
}

func (r *Raft_Implementer) chan_Exit_Inbox() chan *bool {
	return r.chan_exit_inbox
}
func (r *Raft_Implementer) chan_Exit_Follower() chan *bool {
	return r.chan_exit_follower
}
func (r *Raft_Implementer) chan_Exit_Leader() chan *bool {
	return r.chan_exit_leader
}
func (r *Raft_Implementer) chan_Exit_Candidate() chan *bool {
	return r.chan_exit_candidate
}


func (r *Raft_Implementer) chan_Follower() chan *Raft_Msg_Envelope {
	return r.chan_of_follower
}

func (r *Raft_Implementer) chan_Leader() chan *Raft_Msg_Envelope {
	return r.chan_of_leader
}
func (r *Raft_Implementer) chan_Candidate() chan *Raft_Msg_Envelope {
	return r.chan_of_candidate
}





func (r *Raft_Implementer) send_To_OutBox(e *cluster.Envelope) {
	r.S.Outbox()<-e
}
func (r *Raft_Implementer) Recv_From_Inbox() {
	for {
		select {
		case e:=<-r.S.Inbox() :



			msg, _ := e.Msg.([]byte)
			msg_buffer:=bytes.NewBuffer(msg)
			dec:=gob.NewDecoder(msg_buffer)
			var v Raft_Msg_Envelope
			errr:= dec.Decode(&v)
			if errr != nil {
				//log.Fatal("decode error:", err)
				fmt.Println("\nError in decoding")
			}






			switch r.phase {
			case 0:		//follower


				if v.Type_of_raft_msg==1 {
					if v.Term>r.curr_term {
						r.chan_Follower()<-&v
					}
				} else {
					r.chan_Follower()<-&v
				}

			case 1:		//leader

				if v.Type_of_raft_msg==1 {
					//this cannot be no longer be leader if vote_request term is greater than its own term
					if v.Term>r.curr_term {
						r.chan_Leader()<-&v
					}

				} else {
					//ignore the request
				}
			case 2:		//candidate

				if v.Type_of_raft_msg==1 {
					if v.Term>r.curr_term {
						//this cannot be no longer a candidate so change phase to follower & redirect this request to follower
						b:=true
						r.phase=0
						r.chan_exit_candidate<-&b
						//r.chan_enable_raft_controller<-&b
						r.S.Inbox()<-e
						continue

					} else {
						//ignore the request
					}
				} else {
					r.chan_of_candidate<-&v
				}



			}
		case exit:=<-r.chan_Exit_Inbox() :
			fmt.Printf("\nExiting recv_from_inbox of %d:%v",r.pid,exit)
			return


		}
	}
}
func (r *Raft_Implementer) follower() {
	fmt.Printf("\n%d in follower",r.pid)

	for {

		select {
		case e:=<-r.chan_Follower() :
			//fmt.Printf("\nIn follower received:%v",e)
			switch  e.Type_of_raft_msg {
			case 1:   //vote request
				//if ownterm<other term
				//send vote grant
				//other condition will be taken care in message redirector

				//following condition will always satisfied if vote request comes here

				if r.curr_term < e.Term {
					r.curr_term=e.Term
					//str:=r.encapsulate_raft__msg(2,r.s.Pid(),e.Term,"vote grant")


					var msg_buffer bytes.Buffer
					enc := gob.NewEncoder(&msg_buffer)
					d:=Raft_Msg_Envelope{Type_of_raft_msg:2,Pid:r.S.Pid(),Term:e.Term,Actual_msg:"vote grant"}
					err := enc.Encode(&d)
					if err != nil {
						//log.Fatal("encode error:", err)
						fmt.Printf("\nError in encoding")
					}

					b:=msg_buffer.Bytes()



					j:=r.createMsgToSend(e.Pid,0,b)
					r.send_To_OutBox(&j)
					//fmt.Printf("\nSent Vote granted")

				}

			case 2:	  //vote grant
				//ignore

			case 3:	  //heartbeat
				//ignore

			}

		case <-time.After(time.Duration(r.timeout)*time.Millisecond) :
			//case <-time.After(10000*time.Millisecond) :
			//fmt.Printf("\nTimeout in follower")
			r.phase=2
			//b:=true
			//r.chan_enable_raft_controller<-&b
			return



		case exit:=<-r.chan_Exit_Follower() :
			fmt.Printf("\nExiting from follower of %d:%v",r.pid,exit)

			return

		}


	}
}
func (r *Raft_Implementer) leader() {
	fmt.Printf("\n%d in leader phase",r.pid)
	for {

		select {
		case e:=<-r.chan_Leader() :
			//fmt.Printf("\nIn leader received:%v",e)
			switch  e.Type_of_raft_msg {
			case 1:   //vote request
				//here this condition is always satisfied when vote request comes
				if e.Term>r.curr_term {
					r.curr_term=e.Term
					r.phase=0
					//b:=true
					//r.chan_enable_raft_controller<-&b
					return
				}
			case 2:	  //vote grant
				//ignore vote grants here
			case 3:	  //heartbeat
				//ignore never happen

			}
			//following is the heart beat frequency
		case <-time.After(100*time.Millisecond) :
			//broadcast heart beats
			//str:=r.encapsulate_raft__msg(3,r.s.Pid(),r.curr_term,"Heart beat")

			var msg_buffer bytes.Buffer
			enc := gob.NewEncoder(&msg_buffer)
			d:=Raft_Msg_Envelope{Type_of_raft_msg:3,Pid:r.S.Pid(),Term:r.curr_term,Actual_msg:"Heart beat"}
			err := enc.Encode(&d)
			if err != nil {
				//log.Fatal("encode error:", err)
				fmt.Printf("\nError in encoding")
			}
			b:=msg_buffer.Bytes()





			j:=r.createMsgToSend(-1,10,b)
			r.send_To_OutBox(&j)
			//fmt.Printf("\nHeart Beat broadcasted")


		case exit:=<-r.chan_Exit_Leader() :
			fmt.Printf("Exiting from leader of %d:%v",r.pid,exit)
			return

		}

	}
}
func (r *Raft_Implementer) candidate() {
	vote_granted:=1
	//broadcast vote request
	fmt.Printf("\n%d in candidate phase",r.pid)
	r.curr_term++
	//str:=r.encapsulate_raft__msg(1,r.s.Pid(),r.curr_term,"vote request")


	var msg_buffer bytes.Buffer
	enc := gob.NewEncoder(&msg_buffer)
	d:=Raft_Msg_Envelope{Type_of_raft_msg:1,Pid:r.S.Pid(),Term:r.curr_term,Actual_msg:"vote request"}
	err := enc.Encode(&d)
	if err != nil {
		//log.Fatal("encode error:", err)
		fmt.Printf("\nError in encoding")
	}

	b:=msg_buffer.Bytes()


	j:=r.createMsgToSend(-1,0,b)
	r.send_To_OutBox(&j)
	//fmt.Printf("\nVote request send")

	for {


		select {


		case e:=<-r.chan_Candidate() :
			//fmt.Printf("\nIn candidate received:%v",e)
			switch  e.Type_of_raft_msg {
			case 1:   //vote request
				//if r.curr_term<e.term this condition will be handled by code distributing the messages
				//else ignore the request
			case 2:	  //vote grant
				vote_granted++
				//ignore vote grants in leader phase
				if vote_granted >len(r.S.Peers())/2 {
					r.phase=1
					//b:=true
					//r.chan_enable_raft_controller<-&b
					return
				}
			case 3:	  //heartbeat



			}
			//case <-time.After(300*time.Millisecond) :
		case <-time.After(time.Duration(r.timeout)*time.Millisecond) :
			//fmt.Printf("\nChanging to follower phase")
			r.phase=0
			//b:=true
			//r.chan_enable_raft_controller<-&b
			r.curr_term++
			return

		case exit:=<-r.chan_Exit_Candidate() :
			fmt.Printf("Exiting from candidate of %d :%v",r.pid,exit)
			return

		}


	}
}
func (r *Raft_Implementer) RaftController() {
	//b:=true
	//var mutex = &sync.Mutex{}
	//for ;*r.is_enable==true; {
	for {
		select {
		case <-r.chan_exit_raft_controller :
			fmt.Printf("\nExiting from raft controller of %d",r.pid)
			return
		default :
			switch r.phase {

			case 0:
				r.follower()

			case 1:
				r.leader()


			case 2:
				r.candidate()

			}

		}


	}
	fmt.Printf("\nExiting")

}
func (r *Raft_Implementer) Close() {
	b:=false
	switch r.phase {
	case 0:
		r.phase=-1
	r.chan_Exit_Follower()<-&b
	case 1:
		r.phase=-1
	r.chan_Exit_Leader()<-&b
	case 2:
		r.phase=-1
	r.chan_Exit_Candidate()<-&b
	}
	r.chan_exit_raft_controller<-&b
	r.S.Chan_exit_send<-&b
	//r.S.chan_exit_receive<-&b
	r.chan_exit_inbox<-&b



}

func NewRaft(myid int, size_of_in_chan int,size_of_out_chan int,delay_before_conn time.Duration,configfile string,parent_dir string) *Raft_Implementer {


	host_id,host_ip,host_port,host_dir,host_timeout,list_serv,host_own_index:=get_serverinfo(myid,configfile)

	new_chan_of_follower:=make(chan *Raft_Msg_Envelope)
	new_chan_of_leader:=make(chan *Raft_Msg_Envelope)
	new_chan_of_candidate:=make(chan *Raft_Msg_Envelope)
	new_chan_exit_follower:=make(chan *bool)
	new_chan_exit_leader:=make(chan *bool)
	new_chan_exit_candidate:=make(chan *bool)
	new_chan_exit_raft_controller:=make(chan *bool)
	new_chan_exit_inbox:=make(chan *bool)



	fnm := parent_dir+"/cluster/serverlist" + strconv.Itoa(host_id) + ".xml"

	my_server_main:= cluster.New(host_id, size_of_in_chan ,size_of_out_chan,fnm,delay_before_conn)
	raft=Raft_Implementer{S:my_server_main,pid:host_id,ip:host_ip,port:host_port,dir:host_dir,timeout:host_timeout,serv:list_serv,own_index:host_own_index,curr_term:0,heart_beat_frequency:3000*time.Millisecond,chan_of_follower:new_chan_of_follower,chan_of_leader:new_chan_of_leader,chan_of_candidate:new_chan_of_candidate,chan_exit_follower:new_chan_exit_follower,chan_exit_leader:new_chan_exit_leader,chan_exit_candidate:new_chan_exit_candidate,chan_exit_raft_controller:new_chan_exit_raft_controller,chan_exit_inbox:new_chan_exit_inbox,phase:0}









	return &raft



}

type Raftserverinfo struct {
	XMLName    Raftserverlist `xml:"raftserverinfo"`
	Raftserverlist Raftserverlist `xml:"raftserverlist"`
}

type Raftserverlist struct {
	Raftserver []Raftserver `xml:"raftserver"`
}
type Raftserver struct {
	Id   int `xml:"id"`
	Ip   string `xml:"ip"`
	Port int `xml:"port"`
	Raft RaftStore `xml:"raft"`
}
type RaftStore struct {
	Dir string `xml:"dir"`
	Electiontimeout int `xml:"electiontimeout"`
}

/*max char of xmlfile are 100000 words*/
func get_serverinfo(id int, fnm string) (int, string, int, string, int,[]Raftserver, int) {


	xmlFile, err := os.Open(fnm)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return 0, "0", 0,"0",0,nil, -1
	}
	defer xmlFile.Close()
	data := make([]byte, 100000)
	count, err := xmlFile.Read(data)
	if err != nil {
		fmt.Println("Can't read the data", count, err)
		return 0, "0", 0,"0",0,nil, -1
	}
	var q Raftserverinfo
	xml.Unmarshal(data[:count], &q)
	checkError(err)


	for k, sobj := range q.Raftserverlist.Raftserver {
		if sobj.Id==id {
			return q.Raftserverlist.Raftserver[k].Id, q.Raftserverlist.Raftserver[k].Ip, q.Raftserverlist.Raftserver[k].Port,q.Raftserverlist.Raftserver[k].Raft.Dir,q.Raftserverlist.Raftserver[k].Raft.Electiontimeout,q.Raftserverlist.Raftserver, k

		}
	}

	return 0, "0", 0,"0",0,nil, -1
}
func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

