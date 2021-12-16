package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	// "io/ioutil"

	"badcoin/src/node"

	"github.com/urfave/cli"
)

func HealthCheck(c *cli.Context) error {
	var res node.HealthCheckResponse
	err := Get("", &res)
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// SendTx <to address> <amount> -from=<from address> -memo=<some data>
func SendTx(c *cli.Context) error {
	if len(c.Args()) < 2 {
		return fmt.Errorf("To and amount must be specified")
	}
	to := c.Args()[0]
	value := c.Args()[1]
	data := c.Args()[2]

	fmt.Println("sending", value, "to", to, "...")
	var res node.SendTxResponse
	err := Call("tx/send", map[string]string{
		"to":    to,
		"value": value,
		"data":  data,
	}, &res)

	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func SendSignedTx(c *cli.Context) error {
	if len(c.Args()) < 4{
		return fmt.Errorf("to, amount, pubkey and signature must be specified")
	}
	to := c.Args()[0]
	value := c.Args()[1]
    pubkey64 := c.Args()[2]
    signature64 := c.Args()[3]
	data := c.Args()[4]

	fmt.Println("sending signed tx ", value, " BDC to", to, "...")
	var res node.SendTxResponse
	err := Call("tx/send", map[string]string{
		"to":    to,
		"value": value,
        "pubKey": pubkey64,
        "signature": signature64,
		"data":  data,
	}, &res)

	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func NewAddress(c *cli.Context) error {
	var res node.NewAddressResponse
	err := Call("address/new", map[string]string{}, &res)
	if err != nil {
		return err
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func GetInfo(c *cli.Context) error {
	var res node.GetInfoResponse
	err := Get("info", &res)
	if err != nil {
		return err
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func Call(cmd string, options map[string]string, out interface{}) error {
	vals := make(url.Values)
	for k, v := range options {
		vals.Set(k, v)
	}
	resp, err := http.PostForm("http://127.0.0.1:3000/"+cmd, vals)

	//fmt.Println("Response:", resp)
	if err != nil {
		return err
	}
	// buf, err := ioutil.ReadAll(resp.Body)
	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		return err
	}
	// fmt.Printl(string(buf))
	return nil
}

func Get(url string, out interface{}) error {
	resp, err := http.Get("http://127.0.0.1:3000/" + url)
	if err != nil {
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(out)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "bdc-cli"
	app.Usage = "rpc client for badcoin"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:    "status",
			Usage:   "shows connection status",
			Aliases: []string{"stat"},
			Action:  HealthCheck,
		},
		{
			Name:    "sendtx",
			Usage:   "send a transaction",
			Aliases: []string{"tx"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "to",
					Value: "",
					Usage: "to address",
				},
				cli.StringFlag{
					Name:  "value",
					Value: "",
					Usage: "BDC amount",
				},
				cli.StringFlag{
					Name:  "data",
					Value: "",
					Usage: "add data to transaction",
				},
			},
			Action: SendTx,
		},
        {
			Name:    "sendsignedtx",
			Usage:   "send a signed transaction",
			Aliases: []string{"stx"},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "to",
					Value: "",
					Usage: "to address",
				},
				cli.StringFlag{
					Name:  "value",
					Value: "",
					Usage: "BDC amount",
				},
                cli.StringFlag{
					Name:  "pubkey",
					Value: "",
					Usage: "base64 of pubkey",
				},
				cli.StringFlag{
					Name:  "signature",
					Value: "",
					Usage: "base64 of signature",
				},
				cli.StringFlag{
					Name:  "data",
					Value: "",
					Usage: "add data to transaction",
				},
			},
			Action: SendSignedTx,
		},
		{
			Name:    "newaddress",
			Usage:   "get new address",
			Aliases: []string{"addr"},
			Action:  NewAddress,
		},
		{
			Name:    "info",
			Usage:   "shows blockchain information",
			Aliases: []string{"i"},
			Action:  GetInfo,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
