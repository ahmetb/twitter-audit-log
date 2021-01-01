package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"facette.io/natsort"
	"github.com/dghubble/oauth1"
)

var cmds = map[string]func(*http.Client) error{
	"id":        printID,
	"following": printFollowing,
	"followers": printIDs("https://api.twitter.com/1.1/followers/ids.json?count=5000&stringify_ids=true"),
	"mutes":     printIDs("https://api.twitter.com/1.1/mutes/users/ids.json?count=5000&stringify_ids=true"),
	"blocks":    printIDs("https://api.twitter.com/1.1/blocks/ids.json?count=5000&stringify_ids=true"),
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
			panic(*v + " env var not set")
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

func apiCall(c *http.Client, v interface{}, endpoint string) error {
	resp, err := c.Get(endpoint)
	if err != nil {
		return fmt.Errorf("cannot make request to %s: %w", endpoint, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return mkRespErr(resp)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func mkRespErr(resp *http.Response) error {
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		resp.Body.Close()
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(b))
	return fmt.Errorf("request (%s) failed (%d %s)\nbody=%s", resp.Request.RequestURI,
		resp.StatusCode, resp.Status,
		string(b))
}

func selfID(c *http.Client) (string, error) {
	var v struct {
		ID string `json:"id_str"`
	}
	err := apiCall(c, &v, "https://api.twitter.com/1.1/account/verify_credentials.json")
	return v.ID, err
}

type followListResponseV2 struct {
	Data []struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

func printID(c *http.Client) error {
	id, err := selfID(c)
	if err != nil {
		return fmt.Errorf("failed to get self id: %w", err)
	}
	fmt.Print(id)
	return nil
}

func printFollowing(c *http.Client) error {
	id, err := selfID(c)
	if err != nil {
		return fmt.Errorf("failed to get self id: %w", err)
	}
	out, err := following(c, id)
	if err != nil {
		return err
	}

	// Twitter API returns following list in chronologically reverse order.
	// Print in chronological order here.
	for i := len(out) - 1; i >= 0; i-- {
		fmt.Printf("%s,%s\n", out[i][0], out[i][1])
	}
	return nil
}

func following(c *http.Client, id string) ([][2]string, error) {
	var out [][2]string
	var next string

	origBase := `https://api.twitter.com/2/users/:id/following?max_results=500`
	for {
		base := origBase
		if next != "" {
			base += "&pagination_token=:next"
		}
		var v followListResponseV2
		if err := apiCall(c, &v, urlTemplate(base, "id", id, "next", next)); err != nil {
			return nil, fmt.Errorf("failed to query following: %w", err)
		}
		for _, vv := range v.Data {
			out = append(out, [2]string{vv.ID, vv.Username})
		}
		if v.Meta.NextToken == "" {
			break
		}
		next = v.Meta.NextToken
	}
	return out, nil
}

type idListResponseV1 struct {
	IDs           []string `json:"ids"`
	NextCursorStr string   `json:"next_cursor_str"`
	NextCursor    int64    `json:"next_cursor"`
}

func printIDs(endpoint string) func(*http.Client) error {
	return func(c *http.Client) error {
		id, err := selfID(c)
		if err != nil {
			return fmt.Errorf("failed to get self id: %w", err)
		}
		out, err := listIDsFromEndpoint(c, id, endpoint)
		if err != nil {
			return err
		}

		natsort.Sort(out)
		for _, v := range out {
			fmt.Printf("%s\n", v)
		}
		return nil
	}
}

func listIDsFromEndpoint(c *http.Client, id, endpoint string) ([]string, error) {
	var out []string
	var next string
	for {
		base := endpoint
		if next != "" {
			base += "&cursor=:next"
		}
		var resp idListResponseV1
		if err := apiCall(c, &resp, urlTemplate(base, "next", next)); err != nil {
			return nil, fmt.Errorf("failed to query followers: %w", err)
		}
		out = append(out, resp.IDs...)
		if resp.NextCursor == 0 {
			break
		}
		next = resp.NextCursorStr
	}
	return out, nil
}

// urlTemplate replaces all :k1, :k2 etc occurences in base URL
// with values provided as [k1,v1,k2,v2,...].
func urlTemplate(base string, kvp ...string) string {
	for i := 1; i < len(kvp); i += 2 {
		k, v := kvp[i-1], kvp[i]
		base = strings.ReplaceAll(base, ":"+k, v)
	}
	return base
}
