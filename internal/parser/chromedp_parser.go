package parser

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

// Uses chrome dev tools protocol to parse data
type ChromeParser struct {
	matchers [4]*regexp.Regexp
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewChromeParser() (*ChromeParser, error) {

	titlerx, err := regexp.Compile(`"title-info-title-text.*?<`)
	if err != nil {
		return nil, err
	}

	pricerx, err := regexp.Compile(`js-item-price.*?<`)
	if err != nil {
		return nil, err
	}

	imagerx1, err := regexp.Compile("image-frame-cover.*?<div")
	if err != nil {
		return nil, err
	}
	imagerx2, err := regexp.Compile(`src="(.*?)"`)
	if err != nil {
		return nil, err
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	err = chromedp.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("error booting a browser: %w", err)
	}
	go func() {
		<-chromedp.FromContext(ctx).Browser.LostConnection
		fmt.Println("LOST!")
	}()
	return &ChromeParser{
		matchers: [4]*regexp.Regexp{titlerx, pricerx, imagerx1, imagerx2},
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (p *ChromeParser) Parse(timeout time.Duration, url string) *ParseResult {
	var html string

	// get rid of somehow
	ctx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(c context.Context) error {
			node, err := dom.GetDocument().Do(c)
			if err != nil {
				return fmt.Errorf("parser: could not get document: %w", err)
			}

			res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(c)
			if err != nil {
				return fmt.Errorf("parser: could not get outer html of body: %w", err)
			}

			html = res
			return nil
		}),
	)
	if errors.Is(err, context.Canceled) {
		// Restart the browser and re-do the task
		time.Sleep(time.Millisecond * 200)
		fmt.Println("retrying...")
		return p.retry(timeout, url)
	}

	if err != nil {
		err = fmt.Errorf("parser: parsing error: %w", err)
		return NewParseResultWithError(err)
	}

	title := p.parseTitle(&html)
	price, err := p.parsePrice(&html)
	if err != nil {
		err = fmt.Errorf("parser: price parsing error: %w", err)
		return NewParseResultWithError(err)
	}

	return NewParseResult(title, price, url)
}

func (p *ChromeParser) retry(timeout time.Duration, url string) *ParseResult {
	p.ctx, p.cancel = chromedp.NewContext(context.Background())
	return p.Parse(timeout, url)
}

func (p *ChromeParser) parseTitle(buff *string) string {
	rx := p.matchers[0]

	match := rx.FindAllString(*buff, 1)[0]
	spl := strings.Split(match, "")

	var title string
	for i := len(spl) - 2; i >= 0; i-- {
		if spl[i] == ">" {
			break
		}

		title = spl[i] + title
	}

	return title
}

func (p *ChromeParser) parsePrice(buff *string) (float64, error) {
	rx := p.matchers[1]

	match := rx.FindAllString(*buff, 1)[0]
	spl := strings.Split(match, "")

	var pricestr string
	for i := len(spl) - 2; i >= 0; i-- {
		if spl[i] == ">" {
			break
		}

		// Compare by charcode (leave only numbers)
		if spl[i][0] < 48 || spl[i][0] > 57 {
			continue
		}

		pricestr = spl[i] + pricestr
	}

	pricefloat, err := strconv.ParseFloat(pricestr, 64)
	if err != nil {
		return 0.0, err
	}

	return pricefloat, nil
}
