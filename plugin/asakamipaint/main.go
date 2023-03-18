package asakamipaint

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/wallet"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("asakamipaint", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  true,
		Help:              "- 冬香酱写的aipaint \n用法:\n#春生画图+prompt:{...}+negative_prompt:{...}+steps=..+h=...+w=...+cfg_scale=...+seed=...",
		PrivateDataFolder: "asakamipaint",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.Send("已启用ai画图")
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.Send("已禁用ai画图")
		},
	})
	jsontmp := map[string]interface{}{
		"enable_hr":                            false,
		"denoising_strength":                   0,
		"firstphase_width":                     0,
		"firstphase_height":                    0,
		"hr_scale":                             2,
		"hr_upscaler":                          "string",
		"hr_second_pass_steps":                 0,
		"hr_resize_x":                          0,
		"hr_resize_y":                          0,
		"prompt":                               "",
		"seed":                                 -1,
		"subseed":                              -1,
		"subseed_strength":                     0,
		"seed_resize_from_h":                   -1,
		"seed_resize_from_w":                   -1,
		"sampler_name":                         "Euler a",
		"batch_size":                           1,
		"n_iter":                               1,
		"steps":                                20,
		"cfg_scale":                            7,
		"width":                                512,
		"height":                               512,
		"restore_faces":                        false,
		"tiling":                               false,
		"negative_prompt":                      "",
		"eta":                                  0,
		"s_churn":                              0,
		"s_tmax":                               0,
		"s_tmin":                               0,
		"s_noise":                              1,
		"override_settings_restore_afterwards": true,
		"sampler_index":                        "Euler",
	}

	engine.OnPrefix("#春生画图").SetBlock(false).Handle(func(ctx *zero.Ctx) {

		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)

		text := ctx.Event.Message.ExtractPlainText()
		prefix := regexp.MustCompile("^#春生画图:")

		r_text := prefix.ReplaceAllString(text, "")
		prompt := getindex(r_text, "prompt:{", "}")
		negative_prompt := getindex(r_text, "negative_prompt:{", "}")
		stepss := getindex2(r_text, "steps=", "\\d*")
		hs := getindex2(r_text, "(h|H)=", "\\d*")
		ws := getindex2(r_text, "(w|W)=", "\\d*")
		seeds := getindex2(r_text, "seed=", "[\\d-]*")
		cfg_scales := getindex2(r_text, "(cfg|CFG)_scale=", "[\\d-]*")

		steps, err := strconv.ParseInt(stepss, 0, 8)
		h, err2 := strconv.ParseInt(hs, 0, 16)
		w, err3 := strconv.ParseInt(ws, 0, 16)
		seed, err4 := strconv.ParseInt(seeds, 0, 64)
		cfg_scale, err5 := strconv.ParseFloat(cfg_scales, 64)

		if err != nil {
			//log.Print(steps)
			ctx.SendGroupMessage(gid, message.Text("步数错误，请检查steps="))
			return
		}
		if steps < 1 || steps > 50 {
			ctx.SendGroupMessage(gid, message.Text("步数越界，请检查steps="))
			return
		}
		if err2 != nil || h%64 != 0 {
			//log.Print(h)
			ctx.SendGroupMessage(gid, message.Text("图片高度错误，请检查h="))
			return
		}
		if h < 128 || h > 1024 {
			ctx.SendGroupMessage(gid, message.Text("图片高度越界，请检查h="))
			return
		}
		if err3 != nil || w%64 != 0 {
			//log.Print(w)
			ctx.SendGroupMessage(gid, message.Text("图片宽度错误，请检查w="))
			return
		}
		if w < 128 || w > 1024 {
			ctx.SendGroupMessage(gid, message.Text("图片宽度越界，请检查w="))
			return
		}
		if err4 != nil {
			//log.Print(w)
			ctx.SendGroupMessage(gid, message.Text("seed错误，请检查seed="))
			return
		}
		if err5 != nil || int(cfg_scale*10)%5 != 0 {
			//log.Print(w)
			ctx.SendGroupMessage(gid, message.Text("cfg_scale错误，请检查cfg_scale="))
			return
		}

		cost := int(h / 256 * w / 256 * steps / 20)
		if money < cost {
			ctx.SendGroupMessage(gid, message.Text("余额不足("+fmt.Sprint(money)+"/"+fmt.Sprint(cost)+")"))
		} else {
			wallet.InsertWalletOf(uid, cost)
			ctx.SendGroupMessage(gid, message.Text("消费"+fmt.Sprint(cost)+"八重币，余额"+fmt.Sprint(money)+"八重币"))
		}

		ctx.SendGroupMessage(gid, message.Text("少女祈祷中..."))

		jsontmp["prompt"] = prompt
		jsontmp["negative_prompt"] = negative_prompt
		jsontmp["steps"] = steps
		jsontmp["height"] = h
		jsontmp["width"] = w
		jsontmp["seed"] = seed
		jsontmp["cfg_scale"] = cfg_scale

		jsonbyte, err := json.Marshal(jsontmp)
		if err != nil {
			ctx.SendGroupMessage(gid, message.Text("json错误"))
			return
		}
		jsonpost := bytes.NewBuffer(jsonbyte)
		resp, err2 := http.Post("https://127.0.0.1:7860/sdapi/v1/txt2img", "application/json", jsonpost)
		if err2 != nil {
			ctx.SendGroupMessage(gid, message.Text("post错误"))
			return
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)
		images := data["images"]
		str := fmt.Sprintf("%s", images)
		str = strings.ReplaceAll(str, " ", "")
		str = strings.ReplaceAll(str, "[", "")
		str = strings.ReplaceAll(str, "]", "")
		str = strings.ReplaceAll(str, "\\t", "")
		str = strings.ReplaceAll(str, "\\n", "")
		img, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			ctx.SendGroupMessage(gid, message.Text("图片解析错误"))
			log.Printf("Error decoding2 string:  %s ", err.Error())
		}
		ctx.SendGroupMessage(gid, message.ImageBytes(img))
	})
}

func getindex(text string, head string, tail string) string {

	prefix := regexp.MustCompile(head + ".*" + tail)
	prefix2 := regexp.MustCompile(head + "|" + tail)
	index0 := prefix.FindString(text)
	index := prefix2.ReplaceAllString(index0, "")
	return index
}

func getindex2(text string, head string, tail string) string {

	prefix := regexp.MustCompile(head + tail)
	prefix2 := regexp.MustCompile(head)
	index0 := prefix.FindString(text)
	index := prefix2.ReplaceAllString(index0, "")
	return index
}
