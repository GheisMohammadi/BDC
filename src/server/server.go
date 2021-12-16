package server

import (
	logger "badcoin/src/helper/logger"
	node "badcoin/src/node"
	"context"
	"encoding/json"
	"math/big"
	b64 "encoding/base64"
	//"errors"
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

	//Setup Get Endpoints
	muxRouter.HandleFunc("/", server.HandleHealthCheck).Methods("GET")
	muxRouter.HandleFunc("/info", server.HandleGetInfo).Methods("GET")
	muxRouter.HandleFunc("/block", server.HandleGetBlock).Methods("GET")
	muxRouter.HandleFunc("/genesis", server.HandleGetGenesis).Methods("GET")

	//Setup Post Endpoints
	muxRouter.HandleFunc("/tx/send", server.HandleSendTx).Methods("POST")
	muxRouter.HandleFunc("/tx/signed/send", server.HandleSendSignedTx).Methods("POST")
	muxRouter.HandleFunc("/address/new", server.HandleNewAddress).Methods("POST")

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

func (srv *Server) HandleSendSignedTx(w http.ResponseWriter, r *http.Request) {

	pubKeystr := r.FormValue("pubKey")
	to := r.FormValue("to")
	val := r.FormValue("value")
	signaturestr64 := r.FormValue("signature")
	data := r.FormValue("data")

	logger.Info("call sendtx ", val, " BDC to", to)

	value, ok := big.NewFloat(0).SetString(val)
	if ok == false {
		panic("invalid tx value")
	}

    signaturestr, _ := b64.StdEncoding.DecodeString(signaturestr64)


	wallet := srv.Node.GetWallet()
	pubKey := []byte(pubKeystr)
	nonce := wallet.Nonce + 1
	signature := []byte(signaturestr)

	v, _ := value.Float64()
	tx := transaction.NewSignedTransaction(pubKey, nonce, to, v, signature, data)

	resp := srv.Node.SendTransaction(tx)
	if resp == nil {
		json.NewEncoder(w).Encode("tx send failed")
		return
	}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		panic(err)
	}
}

func (srv *Server) HandleSendTx(w http.ResponseWriter, r *http.Request) {

	to := r.FormValue("to")
	val := r.FormValue("value")
	data := r.FormValue("data")

	logger.Info("call sendtx ", val, " BDC to", to)

	value, ok := big.NewFloat(0).SetString(val) //strconv.Atoi(val)
	// amt, err := strconv.ParseInt(amount, 10, 64)
	if ok == false {
		panic("invalid tx value")
	}

	wallet := srv.Node.GetWallet()
	pubKey := wallet.PublicKey
	nonce := wallet.Nonce + 1
	//addr := wallet.GetStringAddress()
	// if addr != from {
	// 	panic(errors.New("no access to this wallet address"))
	// }
	//srv.Node.SendTransaction()
	v, _ := value.Float64()
	tx := transaction.NewTransaction(pubKey, nonce, to, v, data)
	tx.Sign(wallet.PrivateKey)

	resp := srv.Node.SendTransaction(tx)
	if resp != nil {
		srv.Node.GetWalletSet().AddMinerNonce()
	}
	err := json.NewEncoder(w).Encode(resp)
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

func (srv *Server) HandleGetBlock(w http.ResponseWriter, r *http.Request) {
	logger.Info("Call getblock")
	qh, ok := r.URL.Query()["height"]
	if !ok {
		panic("invalid height")
	}

	qhi, errConversion := strconv.Atoi(qh[0])
	if errConversion != nil {
		panic("invalid height")
	}

	height := uint64(qhi)

	data, errGetBlock := srv.Node.GetBlock(height)
	if errGetBlock != nil {
		panic(errGetBlock)
	}
	err := json.NewEncoder(w).Encode(data)
	// res, err := json.Marshal(node.GetInfo())
	if err != nil {
		panic(err)
	}
	// fmt.Fprintf(w, string(res))
}

func (srv *Server) HandleGetGenesis(w http.ResponseWriter, r *http.Request) {
	data, errGetBlock := srv.Node.GetBlock(0)
	if errGetBlock != nil {
		panic(errGetBlock)
	}
	err := json.NewEncoder(w).Encode(data)
	// res, err := json.Marshal(node.GetInfo())
	if err != nil {
		panic(err)
	}
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
