package server

import (
	logger "badcoin/src/helper/logger"
	node "badcoin/src/node"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	mux "github.com/gorilla/mux"

	"badcoin/src/transaction"
)

type Server struct {
	Host    string
	Port    string
	Addr    string
	Node    *node.Node
	Miner   bool
	Handler *http.Handler
}

type Message struct {
	Text string
}

// Create Handler
func MakeMuxRouter(server *Server) http.Handler {

	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", server.HandleHealthCheck).Methods("GET")
	muxRouter.HandleFunc("/tx/send", server.HandleSendTx).Methods("POST")
	muxRouter.HandleFunc("/info", server.HandleGetInfo).Methods("GET")
	muxRouter.HandleFunc("/address/new", server.HandleSendTx).Methods("POST")

	return muxRouter
}

func CreateNewServer(ctx context.Context, servernode *node.Node, port string) *http.Server {
	var server Server
	server.Port = port
	server.Host = ""
	server.Addr = server.Host + ":" + server.Port
	server.Node = servernode

	logger.Info("Starting http server")
	mux := MakeMuxRouter(&server)
	logger.Info("Listening on ", port)
	server.Handler = &mux

	hs := &http.Server{
		Addr:           ":" + port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return hs

}

func (srv *Server) HandleSendTx(w http.ResponseWriter, r *http.Request) {
	from := r.FormValue("from")
	to := r.FormValue("to")
	amount := r.FormValue("amount")
	memo := r.FormValue("memo")

	logger.Info("call sendtx", from, to, amount, memo)

	amt, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		panic(err)
	}

	tx := transaction.Transaction{
		Sender:   from,
		Receiver: to,
		Amount:   uint64(amt),
		Memo:     memo,
	}

	err = json.NewEncoder(w).Encode(srv.Node.SendTransaction(&tx))
	if err != nil {
		panic(err)
	}
}

func (srv *Server) HandleGetInfo(w http.ResponseWriter, r *http.Request) {
	logger.Info("Call getinfo")

	err := json.NewEncoder(w).Encode(srv.Node.GetInfo())
	// res, err := json.Marshal(node.GetInfo())
	if err != nil {
		panic(err)
	}
	// fmt.Fprintf(w, string(res))
}

func (srv *Server) HandleNewAddress(w http.ResponseWriter, r *http.Request) {
	logger.Info("Call getnewaddress")

	err := json.NewEncoder(w).Encode(srv.Node.GetNewAddress())
	// res, err := json.Marshal(node.GetInfo())
	if err != nil {
		panic(err)
	}
	// fmt.Fprintf(w, string(res))
}

func (srv *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Call Health Check")
	msg := Message{
		Text: "Badcoin (BDC) is ok!",
	}
	err := json.NewEncoder(w).Encode(msg)
	if err != nil {
		panic(err)
	}
}
