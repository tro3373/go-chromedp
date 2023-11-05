package main

import (
	"context"

	// "log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func start(c *cli.Context) error {
	ctx, cancel := newContext(true)
	defer cancel()

	err := dlPdf(ctx)
	if err != nil {
		return err
	}
	err = jcom(ctx)
	if err != nil {
		return err
	}
	return nil
}

func newContext(ui bool) (context.Context, context.CancelFunc) {
	if ui {
		baseCtx := context.Background()
		opt := chromedp.DefaultExecAllocatorOptions[:]
		opt = append(opt, chromedp.Flag("headless", false))
		opt = append(opt, chromedp.Flag("disable-gpu", false))
		opt = append(opt, chromedp.Flag("enable-automation", false))
		opt = append(opt, chromedp.Flag("disable-extensions", false))
		opt = append(opt, chromedp.Flag("hide-scrollbars", false))
		opt = append(opt, chromedp.Flag("mute-audio", false))
		// opt = append(opt, chromedp.Flag("user-data-dir", "./data"))
		opt = append(opt, chromedp.UserDataDir("./data"))
		alocCtx, alocCancel := chromedp.NewExecAllocator(baseCtx, opt...)
		ctx, cancel := chromedp.NewContext(
			alocCtx,
			chromedp.WithLogf(log.Printf),
			// chromedp.WithErrorf(log.Printf),
			// chromedp.WithDebugf(log.Printf),
		)
		return ctx, func() {
			alocCancel()
			cancel()
		}
	}
	return chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
}

func jcom(ctx context.Context) error {
	userName := os.Getenv("jcom_username")
	userPass := os.Getenv("jcom_password")
	// log.Info("jcom_password", userPass)
	// if len(userPass) > 0 {
	// 	return nil
	// }
	var res string
	err := chromedp.Run(ctx,
		logAction("Navigating member.jcom", chromedp.Navigate(`https://www.member.jcom.co.jp/frontlogin.do`)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			if e := chromedp.Nodes(`div.login_container`, &nodes, chromedp.AtLeast(0)).Do(ctx); e != nil {
				log.Errorln("Failed to find `div.login_container`", e)
				return e
			}
			if len(nodes) == 0 {
				return nil
			}
			var tasks chromedp.Tasks
			tasks = append(tasks, logAction2(chromedp.WaitVisible, `#wrapper #contents main form`))
			tasks = append(tasks, logAction("Wait(#wrapper #contents main form)", chromedp.WaitVisible(`#wrapper #contents main form`)))
			tasks = append(tasks, logAction("Click(#wrapper #contents main form div.login_container div.login_box button.btn_submit)", chromedp.Click(`#wrapper #contents main form div.login_container div.login_box button.btn_submit`, chromedp.NodeVisible)))
			tasks = append(tasks, logAction("Wait(#field_username)", chromedp.WaitVisible(`#field_username`)))
			tasks = append(tasks, logAction("SetValue(input#username)", chromedp.SetValue(`input#username`, userName)))
			tasks = append(tasks, logAction("SetValue(input#password)", chromedp.SetValue(`input#password`, userPass)))
			tasks = append(tasks, logAction("Sleep(1s)", chromedp.Sleep(1*time.Second)))
			tasks = append(tasks, logAction("Click(button#idpwbtn)", chromedp.Click(`button#idpwbtn`)))
			if e := tasks.Do(ctx); e != nil {
				log.Errorln("Failed to Login", e)
				return e
			}
			return nil
		}),
		// chromedp.Evaluate("window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });", &res),
		logAction("Passed!", nil),
		chromedp.Click(`#wrapper #contents #sub #gNavi li.g03 a`, chromedp.NodeVisible),

		chromedp.WaitVisible(`#wait_infinity`),
	)
	if err != nil {
		return err
	}
	log.Infoln(res)
	return nil
}

func logAction(msg string, act chromedp.Action) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if act == nil {
			log.Infof("==> %s", msg)
			return nil
		}
		log.Infof("==> (sta) %s", msg)
		err := act.Do(ctx)
		log.Infof("==> (end) %s", msg)
		return err
	}
}

func logAction2(act func(), opts ...any) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		log.Infof("==> (sta) %s")
		err := act.Do(opts...)
		log.Infof("==> (end) %s")
		return err
	}
}
func p(a ...interface{}) {
	// fmt.Println(a...)
	log.Infof("%+v", a...)
}

func dlPdf(ctx context.Context) error {
	// capture pdf
	var buf []byte
	if err := chromedp.Run(ctx, printToPDF(`https://www.google.com/`, &buf)); err != nil {
		return err
	}

	if err := os.WriteFile("sample.pdf", buf, 0o644); err != nil {
		return err
	}
	log.Infoln("wrote sample.pdf")
	return nil
}

// print a specific pdf page.
func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
