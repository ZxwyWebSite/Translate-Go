package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	translator "github.com/Conight/go-googletrans"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mholt/archiver/v4"
)

type (
	// 消息结构
	Resp struct {
		Type string `json:"type"` // 信息类型
		Msg  string `json:"msg"`  // 提示消息
		Data any    `json:"data"` // 返回数据
	}
	// 翻译/生草
	TransData struct {
		From  string `json:"from"`  // 源语言
		To    string `json:"to"`    // 译语言
		Text  string `json:"text"`  // 待翻译文本
		Force int    `json:"force"` // 翻译次数
	}
	// 文本转语音
	TTSData struct {
		Lang string `json:"lang"` // 语言
		Text string `json:"text"` // 文本
	}
	// 处理状态
	StatRet struct {
		Num  int    `json:"num"`  // 处理次数
		Text string `json:"text"` // 调试文本
	}
	// 回调
	CallbackRet struct {
		Time string `json:"time"` // 耗时
		Text string `json:"text"` // 结果
	}
	// 站点配置
	siteConfig struct {
		Title       string // 标题
		Keywords    string // 关键词
		Description string // 描述
		Scripts     string // 自定义JS
	}
	// 旧版Api
	PhpResp struct {
		IsOk string `json:"isOk"`
		Msg  string `json:"msg"`
		Text string `json:"text"`
	}
)

var (
	// WS 配置
	socket = websocket.Upgrader{
		HandshakeTimeout: time.Second * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域请求
		},
	}
	// 翻译配置
	transConf = translator.Config{
		Proxy: `socks5://zxwy:2082327995@192.168.10.12:1080`,
	}
	// 语言列表
	i18n = []string{`ar`, `be`, `bg`, `ca`, `cs`, `da`, `de`, `el`, `en`, `es`, `et`, `fi`, `fr`, `hr`, `hu`, `is`, `it`, `iw`, `ja`, `ko`, `lt`, `lv`, `mk`, `nl`, `no`, `pl`, `pt`, `ro`, `ru`, `sh`, `sk`, `sl`, `sq`, `sr`, `sv`, `th`, `tr`, `uk`, `zh`}
	// 站点设置
	siteCfg = &siteConfig{
		Title:       `Zxwy翻译`,
		Keywords:    `Zxwy生草机,Zxwy,子虚乌有,zxwy.tk`,
		Description: `Zxwy生草机，把一段文字变得生草。由谷歌翻译提供支持。——ZxwyWebSiteProject.ZxwyGGC`,
	}
)

//go:embed statics.zip
var staticZip string
var staticFS fs.FS

func init() {
	staticFS = archiver.ArchiveFS{
		Stream: io.NewSectionReader(strings.NewReader(staticZip), 0, int64(len(staticZip))),
		Format: archiver.Zip{},
	}
}

// 前端静态文件处理
func staticHandler() gin.HandlerFunc {
	defunc := func(c *gin.Context) {
		c.Next()
	}
	indexFile, err := staticFS.Open(`index.html`)
	if err != nil {
		log.Print(`静态文件[index.html]不存在，可能会影响首页展示`)
		return defunc
	}
	defer indexFile.Close()
	indexCtx, err := io.ReadAll(indexFile)
	if err != nil {
		log.Print(`静态文件[index.html]读取失败，可能会影响首页展示`)
		return defunc
	}
	// fileServer := http.FileServer(http.FS(staticFS))
	// isExist := func(name string) bool {
	// 	if _, err := staticFS.Open(strings.TrimPrefix(name, `/`)); err != nil {
	// 		return false
	// 	}
	// 	return true
	// }
	return func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, `/`)
		// 跳过 API
		if strings.HasPrefix(path, `api`) || path == `api.php` {
			c.Next()
			return
		}
		// 处理站点信息
		if path == `index.html` || path == `` {
			replace := func(m map[string]string, s string) string {
				for k, v := range m {
					s = strings.ReplaceAll(s, k, v)
				}
				return s
			}
			c.Header(`Content-Type`, `text/html`)
			c.String(200, replace(map[string]string{
				`{siteName}`:   siteCfg.Title,
				`{siteKwd}`:    siteCfg.Keywords,
				`{siteDes}`:    siteCfg.Description,
				`{siteScript}`: siteCfg.Scripts,
			}, string(indexCtx)))
			c.Abort()
			return
		}
		// if !isExist(path) {
		// 	// c.Status(http.StatusNotFound)
		// 	c.Next()
		// 	return
		// }
		// fileServer.ServeHTTP(c.Writer, c.Request)
		// if ctp := mime.TypeByExtension(filepath.Ext(path)); ctp != `` {
		// 	c.Header(`Content-Type`, ctp)
		// }
		// 解决访问源映射文件时的 "seeker can't seek" 问题
		if strings.HasSuffix(path, `.map`) {
			file, err := staticFS.Open(path)
			if err != nil {
				c.Next()
				return
			}
			defer file.Close()
			fileCtx, _ := io.ReadAll(file)
			c.String(200, string(fileCtx))
			c.Abort()
			return
		}
		// 返回静态文件
		c.FileFromFS(path, http.FS(staticFS))
		c.Abort()
	}
}

// Translate 一键翻译 ['源语言','译语言','待翻译文本']['结果','错误']
func translate(from, to, text string) (string, error) {
	t, err := translator.New(transConf).Translate(text, from, to)
	return t.Text, err
}

// GrowGrass 一键生草 ['源语言','译语言','待翻译文本','翻译次数']['结果','信息']
func growGrass(from, to, text string, num int) (string, string) {
	newlang := from
	transpath := newlang
	for i := 0; i < num; i++ {
		oldlang := newlang
		for {
			newlang = i18n[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(39)]
			if !strings.Contains(transpath, newlang) {
				break
			}
		}
		out, err := translate(oldlang, newlang, text)
		if err != nil {
			if i == 0 {
				return ``, err.Error()
			}
			newlang = oldlang
			break
		}
		transpath += `->` + newlang
		text = out
		time.Sleep(time.Millisecond * 300)
	}
	text, err := translate(newlang, to, text)
	if err != nil {
		return ``, err.Error()
	}
	return text, transpath + `->` + to
}

// PHP 兼容版Api
func phpHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		mode := c.Query(`mode`)
		from := c.Query(`from`)
		tolang := c.Query(`to`)
		num := c.Query(`force`)
		text := c.Query(`text`)
		var msg, out string
		if from == `` {
			msg = `缺少源语言(from)参数`
		} else if tolang == `` {
			msg = `缺少译语言(to)参数`
		} else if text == `` {
			msg = `缺少待翻译文本(text)参数`
		} else {
			switch mode {
			case `grass`:
				force, err := strconv.ParseInt(num, 10, 0)
				if err != nil {
					msg = err.Error()
				} else {
					out, msg = growGrass(from, tolang, text, int(force))
				}
			case `trans`:
				var err error
				out, err = translate(from, tolang, text)
				if err != nil {
					msg = err.Error()
				} else {
					msg = from + `->` + tolang
				}
			case `tts`:
				msg = `暂不支持此模式`
			default:
				msg = `未定义的运行模式`
			}
		}
		ok := `ok`
		if out == `` {
			ok = `err`
		}
		c.JSON(200, &PhpResp{
			IsOk: ok,
			Msg:  msg,
			Text: out,
		})
		c.Abort()
	}
}

// WebSocket 新版Api
func wsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建连接
		ws, err := socket.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSONP(http.StatusBadRequest, &Resp{
				Type: `error`,
				Msg:  `CreateSocket: ` + err.Error(),
			})
			c.Abort()
			return
		}
		defer ws.Close()
		// 解析信息
		var ret Resp
		var out, msg string
		var t time.Time
		for {
			err = ws.ReadJSON(&ret)
			t = time.Now()
			if err != nil {
				msg = fmt.Sprint(`ReadJSON: `, err)
				break
			}
			data, err := json.Marshal(ret.Data)
			if err != nil {
				msg = fmt.Sprint(`MarshalJSON: `, err)
				break
			}
			switch ret.Type {
			case `trans`:
				var tr TransData
				err = json.Unmarshal(data, &tr)
				if err != nil {
					msg = fmt.Sprint(`UnmarshalJSON: `, err)
					break
				}
				if tr.From == tr.To {
					msg = `翻译模式下源语言不可与译语言相同`
					break
				}
				out, err = translate(tr.From, tr.To, tr.Text)
				if err != nil {
					msg = err.Error()
				}
			case `grass`:
				var tr TransData
				err = json.Unmarshal(data, &tr)
				if err != nil {
					msg = fmt.Sprint(`UnmarshalJSON: `, err)
					break
				}
				if tr.Force <= 0 {
					msg = `翻译次数非法`
					break
				}
				if tr.Force > 38 {
					msg = fmt.Sprintf(`翻译次数超限(%v > 38)`, tr.Force)
					break
				}
				var f, t, o, p string
				t = tr.From
				p = t
				o = tr.Text
				for i := 0; i < tr.Force; i++ {
					f = t
					for {
						t = i18n[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(39)]
						if !strings.Contains(p, t) {
							break
						}
					}
					s, e := translate(f, t, o)
					if e != nil {
						if i == 0 {
							msg = ``
						}
						break
					}
					p += `->` + t
					o = s
					ws.WriteJSON(&Resp{
						Type: `stat`,
						Msg:  f + `->` + t,
						Data: &StatRet{
							Num:  i + 1,
							Text: s,
						},
					})
					time.Sleep(time.Millisecond * 300)
				}
				out, err = translate(t, tr.To, o)
				if err != nil {
					msg = err.Error()
					break
				}
				msg = p + `->` + tr.To
			case `tts`:
				msg = `暂不支持此功能`
			default:
				msg = `不支持的处理类型`
			}
			if true {
				break
			}
		}
		if out == `` {
			ws.WriteJSON(&Resp{
				Type: `error`,
				Msg:  msg,
				Data: struct{}{},
			})
			c.Abort()
			return
		}
		ws.WriteJSON(&Resp{
			Type: `callback`,
			Msg:  msg,
			Data: &CallbackRet{
				Time: time.Since(t).String(),
				Text: out,
			},
		})
	}
}

func initRouter() *gin.Engine {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{`/api/`})))
	r.Use(staticHandler())
	r.GET(`/api.php`, phpHandler())
	g := r.Group(`/api`)
	{
		g.GET(`/php`, phpHandler())
		g.GET(`/ws`, wsHandler())
	}
	return r
}

func main() {
	const version = `1.0.0-β1`
	fmt.Print(`
  ______ ______ ______ _______ _____ _      ______ ______ ______      ______ ______
 /__  __/ __   / __   / ___  / _____/ /    / __   /__  __/ _____/    / _____/ ___  /
   / / / /__/ / /__/ / /  / / /____/ /    / /__/ /  / / / /___  __  / /___ / /  / /
  / / /   ___/ ___  / /  / /____  / /    / ___  /  / / / ____/ /_/ / //_  / /  / /
 / / / /\ \ / /  / / /  / /____/ / /____/ /  / /  / / / /____     / /__/ / /__/ /
/_/ /_/  \_/_/  /_/_/  /_/______/______/_/  /_/  /_/ /______/    /______/______/
================================================================================
   ` + siteCfg.Title + ` translate-go v` + version + `  https://github.com/ZxwyWebSite/Translate-Go
`)
	// gin.SetMode(gin.ReleaseMode)
	r := initRouter()
	r.Run(`:1009`)
}
