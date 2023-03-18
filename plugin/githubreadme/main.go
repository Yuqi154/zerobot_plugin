package githubreadme

import (
	"regexp"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/go-rod/rod"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var browser = rod.New().MustConnect()

func init() {

	engine := control.Register("githubreadme", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Help:              "获取GitHub链接的readme",
		PrivateDataFolder: "githubreadme",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.Send("已启用")
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.Send("已禁用")
		},
	})
	engine.OnMessage().SetBlock(false).Handle(func(ctx *zero.Ctx) {
		msage := ctx.Event.Message.ExtractPlainText()
		re := regexp.MustCompile(`https?://github\.com/\w+/\w+`)
		match := re.MatchString(msage)
		url := re.FindString(msage)
		if match {

			page := browser.MustPage(url).MustWaitLoad()
			pic := page.MustScreenshotFullPage()

			re := regexp.MustCompile(`github.com/(.*)/(.*)`)
			match := re.FindStringSubmatch(url)
			Owner := match[1]
			Repo := match[2]
			ctx.SendChain(message.Image("https://opengraph.githubassets.com/0/"+Owner+"/"+Repo), message.ImageBytes(pic))
		} else {
			return
		}

	})
}
