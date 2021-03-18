package main
import (
	"bufio"
	"context"
	"net/http"
	"path"
	"fmt"
	"encoding/json"
	"log"
	"golang.org/x/net/websocket"
)

type runner struct {
	ctx         context.Context
	running     bool
	container   string
	clientIn    chan []byte
	clientOut   []chan []byte
}

var runners map[string]*runner

func (rn *runner) loop() {
	rn.running=true
	defer func() { rn.running=false }()
	atc,err := attachContainer(rn.ctx, rn.container)
	if err!=nil {
		return
	}
	incoming := make(chan []byte)
	fin := make (chan bool)
	go func(rd *bufio.Reader) {
		for {
			buffer := make([]byte,65536)
			_, err := rd.Read(buffer)
			if err != nil {
				log.Printf("Incoming error: %v", err)
				break
			}
			incoming <- buffer
		}
		fin <- true
	}(atc.Reader)
	for {
		select {
			case s := <- rn.clientIn:
				atc.Conn.Write(s)
			case s := <- incoming:
				for _,c := range rn.clientOut {
					c <- s
				}
			case <- fin:
				return
		}
	}
}

type runnerResponse struct {
	Status string    `json:"status"`
	Id     string    `json:"id,omitempty"`
	Error  string    `json:"error,omitempty"`
}

type listrunnerResponse struct {
	Ids    []string  `json:"ids"`
}

func init() {
	runners=make(map[string]*runner)
}

func httpRet(w http.ResponseWriter, ret *runnerResponse) {
	jresponse, _ := json.Marshal(ret)
	if ret.Status!="success" {
		http.Error(w, string(jresponse), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w,string(jresponse))
	}
	
}

// @id newRunner
// @tags runner
// @summary Launch a new container for interactive use
// @produce application/json
// @success 200 {object} runnerResponse
// @failure 400
// @failure 500
// @router /run/new [post]
func newRunner(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	c_id, err := startContainer(ctx, "player-immuclient")
	if err != nil {
		httpRet(w, &runnerResponse{Status:"fail", Error:err.Error()})
		return
	}
	rn := &runner{ctx:ctx, container:c_id}
	runners[c_id] = rn
	go rn.loop()
	httpRet(w, &runnerResponse{Status:"success", Id:c_id})
}

// @id listRunners
// @tags runner
// @summary List containers ids of running containers for interactive ouse
// @produce application/json
// @success 200 {object} listrunnerResponse
// @failure 400
// @failure 500
// @router /run/list [get]
func listRunners(w http.ResponseWriter, req *http.Request) {
	ret := &listrunnerResponse{Ids:make([]string,0,len(runners))}
	for i,_ := range runners {
		ret.Ids=append(ret.Ids,i)
	}
	jret, _ := json.Marshal(ret)
	fmt.Fprintf(w,string(jret))
}

// @id closeRunner
// @tags runner
// @summary Halts and delete an interactive container
// @param id path string true "Container id"
// @produce application/json
// @success 200 {object} runnerResponse
// @failure 400 {object} runnerResponse
// @failure 500 {object} runnerResponse
// @router /run/close/{id} [post]
func closeRunner(w http.ResponseWriter, r *http.Request) {
	if r.Method!="POST" {
		httpRet(w, &runnerResponse{Status:"fail", Error:"Invalid method"}) 
		return
	}
	reqId := path.Base(r.URL.Path)
	rn, ok := runners[reqId]
	if !ok {
		log.Printf("Id '%s' not found",reqId)
		httpRet(w, &runnerResponse{Status:"fail", Error:"container id not found"}) 
		return
	}
	if err:=stopContainer(rn.ctx, rn.container); err!=nil {
		httpRet(w, &runnerResponse{Status:"fail", Error:err.Error()}) 
		return
	}
	delete(runners,reqId)
	httpRet(w, &runnerResponse{Status:"success", Id:reqId})
}

func eventRunner(w http.ResponseWriter, r *http.Request) {
	reqId := path.Base(r.URL.Path)
	rn, ok := runners[reqId]
	if !ok {
		log.Printf("Id '%s' not found",reqId)
		httpRet(w, &runnerResponse{Status:"fail", Error:"container id not found"}) 
		return
	}
	s := websocket.Server{Handler: websocket.Handler(func(ws *websocket.Conn) {
		wsRunnerEvents(rn, ws)
		})}
	s.ServeHTTP(w,r)
}

func wsRunnerEvents(rn *runner, ws *websocket.Conn) {
	end := make (chan bool)
	clOut := make (chan []byte)
	rn.clientOut=append(rn.clientOut,clOut)
	defer func() {
		for i,c := range rn.clientOut {
			if c==clOut {
				copy(rn.clientOut[i:], rn.clientOut[i+1:])
				rn.clientOut=rn.clientOut[:len(rn.clientOut)-1]
				break
			}
		}
	}()
	go func() {
		for {
			buf := make([]byte,65536) 
			_,err := ws.Read(buf)
			if err != nil {
				log.Printf("runner error %s",err.Error())
				break
			}
			rn.clientIn <- buf
		}
		end <- true
	}()
	loop: for {
		select {
			case <- end:
				break loop
			case b := <- clOut:
				ws.Write(b)
		}
	}
}
