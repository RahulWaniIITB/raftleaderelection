package Raft

import (
	"testing"
	"strconv"
	"sync"
	"time"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"bytes"
	"path"
	//"fmt"


)





func isLeader(rpc_port int) bool {
	client, err := rpc.Dial("tcp", "localhost" + ":"+strconv.Itoa(rpc_port))

	reply:=-1
	if err != nil {

		return false
	} else {


		I:=10

		errr := client.Call("RaftRPC.Phase", &I, &reply)
		if errr != nil {

			//fmt.Println("\nError comes")
			return false
		}
		client.Close()
	}
	//fmt.Printf("\n")
	//fmt.Printf("Rk="+strconv.Itoa(reply))
	if reply==1 {
		//fmt.Printf("\nReturning true")
		return true
	}
	return false
}



func kill(rpc_port int) {


	client, err := rpc.Dial("tcp", "localhost" + ":"+strconv.Itoa(rpc_port))

	reply:=-1
	if err != nil {
		//fmt.Printf("\n%d closed",rpc_port)
		return
	} else {


		I:=10



		errr := client.Call("RaftRPC.Exit", &I, &reply)
		if errr != nil {

			//fmt.Println("\nError comes in %d",rpc_port)
			return
		}
		client.Close()
	}


}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
func run(id int,port int,wg *sync.WaitGroup) {


	go_exec_path, lookErr := exec.LookPath("go")
	if lookErr != nil {
		panic(lookErr)
	}
	//fmt.Println(""+go_exec_path)


	env := os.Environ()


	curr_path,_:=os.Getwd()
	//fmt.Printf(""+path)
	parent,_:=path.Split(curr_path)
	parent=TrimSuffix(parent,"/")
	//fmt.Println("Parent"+parent)




	cmd:=exec.Command("go","run","main.go",strconv.Itoa(id),strconv.Itoa(port),parent)
	cmd.Env=env
	var out bytes.Buffer
	cmd.Stdout = &out

	var errbuff bytes.Buffer
	cmd.Stderr = &errbuff



	//cmd.Dir="/home/rahul/IdeaProjects/cloud/src"
	//cmd.Path="/usr/bin/go"

	cmd.Dir=parent
	cmd.Path=go_exec_path
	cmd.Start()


	cmd.Wait()




	time.Sleep(500*time.Millisecond)

	wg.Done()
}

func TestMajorityMinority(t *testing.T) {
	wg:=new(sync.WaitGroup)
	wg.Add(5)

	rpc_port:=make([]int,5)



	rpc_port[0]=11110
	rpc_port[1]=11111
	rpc_port[2]=11112
	rpc_port[3]=11113
	rpc_port[4]=11114


	for i:=0;i<5;i++ {
		go run(i,rpc_port[i],wg)
	}

	time.Sleep(10*time.Second)



	no_of_phase_requests:=50



	last_leader:=-1

	leader_cnt:=0
	for j:=0;j<no_of_phase_requests;j++ {
		leader_cnt=0
		for i:=0;i<5;i++ {
			if isLeader(rpc_port[i]) {
				leader_cnt++
				last_leader=i
			}


		}
		//fmt.Printf("\nNumber of leaders:%d",cnt)
		time.Sleep(500*time.Millisecond)
	}

	if last_leader==-1 {
		t.Log(strconv.Itoa(last_leader))
		t.Log("No closed Error")
		t.Fail()
	}


	kill(rpc_port[last_leader])

	time.Sleep(10*time.Second)

	//fmt.Printf("\nAfter deleting %d",last_leader)//1


	leader_cnt=0
	test_success_cnt:=0
	for j:=0;j<no_of_phase_requests;j++ {
		leader_cnt=0
		for i:=0;i<5;i++ {


			if isLeader(rpc_port[i]) {
				leader_cnt++
				last_leader=i
			}


		}

		if leader_cnt==1 {
			test_success_cnt++
		}

		//fmt.Printf("\nNumber of leaders:%d",cnt)
		time.Sleep(500*time.Millisecond)
	}

	if test_success_cnt!=no_of_phase_requests {
		t.Log(strconv.Itoa(test_success_cnt)+":"+strconv.Itoa(no_of_phase_requests))
		t.Log("1 closed Error")
		t.Fail()
	}


	kill(rpc_port[last_leader])

	time.Sleep(10*time.Second)

	//fmt.Printf("\nAfter deleting %d",last_leader)//2


	leader_cnt=0
	test_success_cnt=0
	for j:=0;j<no_of_phase_requests;j++ {
		leader_cnt=0
		for i:=0;i<5;i++ {

			if isLeader(rpc_port[i]) {
				leader_cnt++
				last_leader=i
			}

		}
		if leader_cnt==1{
			test_success_cnt++
		}

		//fmt.Printf("\nNumber of leaders:%d",cnt)
		time.Sleep(500*time.Millisecond)
	}

	if test_success_cnt!=no_of_phase_requests {
		t.Log(strconv.Itoa(test_success_cnt)+":"+strconv.Itoa(no_of_phase_requests))
		t.Log("2 closed error")
		t.Fail()
	}


	kill(rpc_port[last_leader])

	time.Sleep(10*time.Second)

	//fmt.Printf("\nAfter deleting %d",last_leader)//3


	leader_cnt=0
	test_success_cnt=0
	for j:=0;j<no_of_phase_requests;j++ {
		leader_cnt=0
		for i:=0;i<5;i++ {

			if isLeader(rpc_port[i]) {
				leader_cnt++
				last_leader=i
			}


		}
		if leader_cnt==1 {
			test_success_cnt++
		}
		//fmt.Printf("\nNumber of leaders:%d",cnt)
		time.Sleep(500*time.Millisecond)
	}
	if(test_success_cnt!=0) {
		t.Log("3 closed error")
		t.Fail()
	}

	//path,_:=os.Getwd()
	//fmt.Printf(""+path)







	//fmt.Println("\nEnd of my test")

}

