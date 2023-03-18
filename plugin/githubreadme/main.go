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
		//log.Print("Loading")
		msage := ctx.Event.Message.ExtractPlainText()
		re := regexp.MustCompile(`https?://github\.com/\w+/\w+`)
		match := re.MatchString(msage)
		url := re.FindString(msage)
		if match {

			//log.Print("Loading")
			pic := getpic(url)

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

func getpic(url string) []byte {
	// 创建一个浏览器实例
	page := browser.MustPage(url).MustWaitLoad()

	// 转到指定的网页

	//log.Print("Pageget")
	// 获取要渲染的div对象
	//pic, err := page.MustElement("#readme").Screenshot("png", 100)
	//if err != nil {
	//	error.Error(err)
	//}
	pic := page.MustScreenshotFullPage()

	//log.Print("gotpic")

	return pic
}
