package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dghubble/oauth1"
)

var cmds = map[string]func(*http.Client) error{
	"me":        me,
	"followers": nil,
	"following": nil,
}

func init() { flag.Parse() }

func main() {
	if flag.NArg() != 1 {
		panic(fmt.Sprintf("requires 1 positional argument (cmd name); got %d args", flag.NArg()))
	}
	cmd, ok := cmds[flag.Arg(0)]
	if !ok {
		panic("unknown command: " + flag.Arg(0))
	}

	var (
		consumerKey    = "TWITTER_CONSUMER_KEY"
		consumerSecret = "TWITTER_CONSUMER_SECRET"
		accessToken    = "TWITTER_ACCESS_TOKEN"
		tokenSecret    = "TWITTER_TOKEN_SECRET"
	)
	for _, v := range []*string{&consumerKey, &consumerSecret, &accessToken, &tokenSecret} {
		if vv := os.Getenv(*v); vv == "" {
			panic(*v + "env var not set")
		} else {
			*v = vv
		}
	}
	ctx := context.Background()
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, tokenSecret)
	client := config.Client(ctx, token)

	if err := cmd(client); err != nil {
		panic(err)
	}
}

func me(c *http.Client) error {
	resp, err := c.Get("https://api.twitter.com/1.1/account/verify_credentials.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return mkRespErr(resp)
	}
	fmt.Println(resp)
	return nil
}

func mkRespErr(resp *http.Response) error {
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		resp.Body.Close()
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	return fmt.Errorf("response failed, status:%d %s\nbody=%s", resp.StatusCode, resp.Status,
		string(b))
}
