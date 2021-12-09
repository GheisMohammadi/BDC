package main 

import (
    "fmt"
    "log"
    "os"
    "net/url"
    "net/http"
    "encoding/json"
    // "io/ioutil"
    
    "github.com/urfave/cli"
    "badcoin/src/node"
)

func HealthCheck(c *cli.Context) error{
    var res node.HealthCheckResponse
    err := Get("", &res)
    out, err := json.MarshalIndent(res, "","  ")
    if err != nil {
        return err
    }
    fmt.Println(string(out))
    return nil
}

// SendTx <to address> <amount> -from=<from address> -memo=<some data>
func SendTx(c *cli.Context) error {
    if len(c.Args()) != 2 {
        return fmt.Errorf("To and amount must be specified")
    }
    to := c.Args()[0]
    amount := c.Args()[1]
    var from string 
    if c.String("from") == "" {
        from = "default"
    } else {
        from = c.String("from")
    }
    memo := c.String("memo")
    
    var res node.SendTxResponse
    err := Call("tx/send", map[string]string{
        "to": to, 
        "amount": amount,
        "from": from,
        "memo": memo,
    }, &res)
    
    out, err := json.MarshalIndent(res, "","  ")
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
    out, err := json.MarshalIndent(res, "","  ")
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
    out, err := json.MarshalIndent(res, "","  ")
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
    resp, err := http.PostForm("http://127.0.0.1:3000/" + cmd, vals)

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
        Name:   "status",
        Usage:  "shows connection status",
        Action: HealthCheck,
      },
      {
        Name:    "sendtx",
        Usage:   "send a transaction",
        Flags: []cli.Flag {
            cli.StringFlag{
                Name: "from",
                Value: "",
                Usage: "from address",
            },
            cli.StringFlag{
                Name: "data",
                Value: "",
                Usage: "add data to transaction",
            },
        },
        Action:  SendTx,
      },
      {
        Name:   "newaddress",
        Usage:  "get new address",
        Action: NewAddress,
      },
      {
        Name:   "info",
        Usage:  "shows blockchain information",
        Action: GetInfo,
      },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
