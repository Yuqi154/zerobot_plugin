package asakamigpt

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/otiai10/openaigo"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("asakamigpt", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Help:              "- 冬香酱写的多线程chatgpt&bing回复 ",
		PrivateDataFolder: "asakamigpt",
	})

	var threada = true
	conna, err1 := net.Dial("tcp", "127.0.0.1:9000")
	if err1 != nil {
		log.Print(err1)
	}

	engine.OnPrefix("testGPT").SetBlock(false).Handle(func(ctx *zero.Ctx) {

		var gid = ctx.Event.GroupID
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)

		ctx.SendChain(message.Text("少女祈祷中"))

		prefix := regexp.MustCompile("^testGPT([:：])?")
		text := ctx.Event.Message.ExtractPlainText()
		r_text := prefix.ReplaceAllString(text, "")

		log.Print("[gptapi] loading")

		resa := testgpt(r_text)

		if len(resa) > 50 {
			cost := len(resa) / 50
			if money < cost {
				//ctx.SendGroupMessage(gid, message.Text("余额不足("+fmt.Sprint(money)+"/"+fmt.Sprint(cost)+")"))
			} else {
				//wallet.InsertWalletOf(uid, cost)
			}
		}

		ctx.SendGroupMessage(gid, message.Text(resa))

		log.Print("[gptapi] exited")
	})

	engine.OnPrefix("GPT").SetBlock(false).Handle(func(ctx *zero.Ctx) {

		var gid = ctx.Event.GroupID
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)

		ctx.SendChain(message.Text("少女祈祷中"))

		prefix := regexp.MustCompile("^GPT([:：])?")
		text := ctx.Event.Message.ExtractPlainText()
		r_text := prefix.ReplaceAllString(text, "")

		log.Print("[gptapi] loading")

		resa := gpt(r_text)

		if len(resa) > 50 {
			cost := len(resa) / 50
			if money < cost {
				ctx.SendGroupMessage(gid, message.Text("余额不足("+fmt.Sprint(money)+"/"+fmt.Sprint(cost)+")"))
			} else {
				wallet.InsertWalletOf(uid, cost)
			}
		}

		ctx.SendGroupMessage(gid, message.Text(resa))

		log.Print("[gptapi] exited")
	})

	engine.OnPrefix("Bing").SetBlock(false).Handle(func(ctx *zero.Ctx) {

		var gid = ctx.Event.GroupID
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)
		if threada {

			ctx.SendChain(message.Text("少女祈祷中"))

			prefix := regexp.MustCompile("^[bB][iI][nN][gG]([:：])?")
			text := ctx.Event.Message.ExtractPlainText()
			r_text := prefix.ReplaceAllString(text, "")

			threada = false

			log.Print("[bingapi] loading")

			resa := api(r_text, conna)

			if len(resa) > 50 {
				cost := len(resa) / 50
				if money < cost {
					ctx.SendGroupMessage(gid, message.Text("余额不足("+fmt.Sprint(money)+"/"+fmt.Sprint(cost)+")"))
				} else {
					wallet.InsertWalletOf(uid, cost)
				}
			}

			ctx.SendGroupMessage(gid, message.Text(resa))

			threada = true
			log.Print("[bingapi] exited")

		} else {
			log.Print("[bingapi] buzy")
			ctx.SendGroupMessage(gid, message.Text("bing正忙"))
		}
	})

	engine.OnMessage(zero.OnlyToMe).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		var gid = ctx.Event.GroupID
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)
		if threada {
			r_text := ctx.Event.Message.ExtractPlainText()

			threada = false

			log.Print("[bingapi] loading")

			resa := api(r_text, conna)

			if len(resa) > 50 {
				cost := len(resa) / 50
				if money < cost {
					ctx.SendGroupMessage(gid, message.Text("余额不足("+fmt.Sprint(money)+"/"+fmt.Sprint(cost)+")"))
				} else {
					wallet.InsertWalletOf(uid, cost)
				}
			}

			ctx.SendGroupMessage(gid, message.Text(resa))

			threada = true

			log.Print("[bingapi] exited")
		} else {
			log.Print("[bingapi] buzy")
			ctx.SendGroupMessage(gid, message.Text("bing正忙"))
		}
	})
}

func api(in string, conn net.Conn) string {

	conn.Write([]byte(in))
	conn.SetReadDeadline(time.Now().Add(200 * time.Second))
	buf := make([]byte, 4096)

	for {
		_, err := conn.Read(buf)
		if strings.Contains(string(buf[:]), "respont:") {
			break
		}
		if err != nil {
			return "connection panic"
		}
	}

	prefix := regexp.MustCompile(".*respont:")
	out := string(buf[:])
	out = prefix.ReplaceAllString(out, "")

	return out
}

func gpt(in string) string {

	client := openaigo.NewClient([]string{
		"sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	}[rand.Intn(3)])

	// You can set whatever you want
	url := Proxyurl()
	transport := &http.Transport{
		Proxy: http.ProxyURL(url),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client.HTTPClient = &http.Client{Transport: transport}
	// Done!

	request := openaigo.ChatCompletionRequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []openaigo.ChatMessage{
			{Role: "user", Content: in},
		},
	}
	ctx := context.Background()
	response, err := client.Chat(ctx, request)
	if err != nil {
		log.Print(err)
	}
	return response.Choices[0].Message.Content
}

func testgpt(in string) string {

	client := openaigo.NewClient([]string{
		"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",
		"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",
		"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",
	}[rand.Intn(3)])

	// You can set whatever you want
	//transport := &http.Transport{}
	//client.HTTPClient = &http.Client{Transport: transport}
	// Done!

	request := openaigo.ChatCompletionRequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []openaigo.ChatMessage{
			{Role: "user", Content: in},
		},
	}
	ctx := context.Background()
	response, err := client.Chat(ctx, request)
	if err != nil {
		log.Print(err)
	}
	return response.Choices[0].Message.Content
}

func Proxyurl() *url.URL {
	urla := []string{
		"https://49.212.143.246:6666",
		"https://8.209.253.237:8080",
		"https://8.209.249.96:2080",
		"https://8.209.243.173:8080",
		"https://47.245.34.161:2020",
		"https://8.209.240.66:2080",
		"https://8.209.253.237:443",
		"https://8.209.243.173:13",
		"https://47.245.34.161:9999",
		"https://8.209.243.173:80",
		"https://8.209.249.96:9199",
		"https://47.245.34.161:8080",
		"https://138.2.55.182:8080",
		"https://47.245.34.161:20201",
		"https://8.209.249.96:7302",
	}[rand.Intn(15)]
	log.Printf("Use proxy: " + urla)
	proxyURL, err := url.Parse(urla)
	if err != nil {
		log.Print(err)
	}
	return proxyURL
}
